// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

// Clock example - publishes or subscribes to UTC timestamps.
//
// This example demonstrates how to use qgo for publishing and subscribing.
// It can operate in two modes:
//   - publish: Publishes the current UTC time as objects once per second
//   - subscribe: Subscribes to a clock track and displays received timestamps
//
// The publish mode supports two flows:
//   - announce flow (use_announce=true): Announces namespace, waits for subscribers
//   - direct flow (use_announce=false): Publishes directly without announce
//
// Usage:
//
//	clock -mode publish -server localhost:4433 -namespace clock/utc -track time
//	clock -mode subscribe -server localhost:4433 -namespace clock/utc -track time
//	clock -mode publish -use_announce -server localhost:4433
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/quicr/qgo"
)

func main() {
	server := flag.String("server", "localhost:4433", "Relay server address (host:port)")
	transport := flag.String("transport", "quic", "Transport protocol (quic or webtransport)")
	namespace := flag.String("namespace", "clock/utc", "Namespace for clock data")
	track := flag.String("track", "time", "Track name")
	priority := flag.Uint("priority", 128, "Priority (0-255)")
	ttl := flag.Uint("ttl", 5000, "TTL in milliseconds")
	mode := flag.String("mode", "publish", "Mode: publish or subscribe")
	useAnnounce := flag.Bool("use_announce", false, "Use announce flow (publish mode only)")
	flag.Parse()

	// Parse transport type
	var transportType qgo.Transport
	switch *transport {
	case "quic":
		transportType = qgo.TransportQUIC
	case "webtransport", "wt":
		transportType = qgo.TransportWebTransport
	default:
		log.Fatalf("Unknown transport: %s (use 'quic' or 'webtransport')", *transport)
	}

	// Create client configuration
	cfg := qgo.ClientConfig{
		ConnectURI: *server,
		EndpointID: "qgo-clock-" + *mode,
		Transport:  transportType,
	}

	log.Printf("Creating client with transport: %s, server: %s", transportType, *server)

	// Create client
	client, err := qgo.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Set up status change callback
	client.OnStatusChange(func(status qgo.ClientStatus) {
		log.Printf("Client status: %s", status)
	})

	// Connect to relay
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Connecting to %s...", *server)
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	log.Println("Connected!")

	// Parse namespace and create full track name
	ns := qgo.ParseNamespace(*namespace)
	ftn := qgo.FullTrackName{
		Namespace: ns,
		TrackName: qgo.NewTrackName(*track),
	}

	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	switch *mode {
	case "publish":
		runPublisher(client, ns, ftn, *useAnnounce, uint8(*priority), uint32(*ttl), sigCh)
	case "subscribe":
		runSubscriber(client, ftn, uint8(*priority), sigCh)
	default:
		log.Fatalf("Unknown mode: %s (use 'publish' or 'subscribe')", *mode)
	}
}

func runPublisher(client *qgo.Client, ns qgo.Namespace, ftn qgo.FullTrackName, useAnnounce bool, priority uint8, ttl uint32, sigCh chan os.Signal) {
	var nsHandler *qgo.PublishNamespaceHandler
	// If using announce flow, announce the namespace first
	if useAnnounce {
		log.Printf("Announcing namespace: %s", ns.String())
		var err error
		nsHandler, err = qgo.NewPublishNamespaceHandler(ns)
		if err != nil {
			log.Fatalf("Failed to create namespace handler: %v", err)
		}
		defer nsHandler.Close()

		if err := client.PublishNamespace(nsHandler); err != nil {
			log.Fatalf("Failed to publish namespace: %v", err)
		}
		log.Println("Namespace announced, waiting for subscribers...")
	} else {
		log.Println("Using direct publish flow (no announce)")
	}

	// Create publish track configuration
	pubCfg := qgo.PublishTrackConfig{
		FullTrackName: ftn,
		TrackMode:     qgo.TrackModeStream,
		Priority:      priority,
		TTL:           ttl,
		UseAnnounce:   useAnnounce,
	}

	// Create publish handler
	handler, err := client.PublishTrack(pubCfg)
	if err != nil {
		log.Fatalf("Failed to create publish handler: %v", err)
	}
	defer handler.Close()

	// Set up status change callback
	handler.OnStatusChange(func(status qgo.PublishStatus) {
		log.Printf("Publish status: %s", status)
	})

	// Publishing loop
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var groupID, objectID uint64
	const objectsPerGroup = 15

	log.Printf("Publishing UTC timestamps to %s...", ftn.String())
	log.Println("Press Ctrl+C to stop")

	for {
		select {
		case <-sigCh:
			log.Println("Shutting down...")
			return

		case <-ticker.C:
			if !handler.CanPublish() {
				log.Printf("Cannot publish yet (status: %s)", handler.Status())
				continue
			}

			// Increment IDs
			objectID++
			if objectID > objectsPerGroup {
				objectID = 1
				groupID++
			}

			// Create timestamp payload
			timestamp := time.Now().UTC().Format(time.RFC3339Nano)
			data := []byte(timestamp)

			// Create headers
			headers := qgo.ObjectHeaders{
				GroupID:  groupID,
				ObjectID: objectID,
				Priority: priority,
			}

			// Publish
			status, err := handler.PublishObject(headers, data)
			if err != nil {
				log.Printf("Publish error: %v", err)
			} else if status.IsSuccess() {
				log.Printf("Published [%d:%d] %s", groupID, objectID, timestamp)
			} else {
				log.Printf("Publish status: %s", status)
			}
		}
	}
}

func runSubscriber(client *qgo.Client, ftn qgo.FullTrackName, priority uint8, sigCh chan os.Signal) {
	// Create subscribe track configuration
	subCfg := qgo.SubscribeTrackConfig{
		FullTrackName: ftn,
		Priority:      priority,
		GroupOrder:    qgo.GroupOrderAscending,
		FilterType:    qgo.FilterTypeLatestObject,
	}

	// Create subscribe handler
	handler, err := client.SubscribeTrack(subCfg)
	if err != nil {
		log.Fatalf("Failed to create subscribe handler: %v", err)
	}
	defer handler.Close()

	// Set up status change callback
	handler.OnStatusChange(func(status qgo.SubscribeStatus) {
		log.Printf("Subscribe status: %s", status)
	})

	// Set up object received callback
	handler.OnObjectReceived(func(obj qgo.Object) {
		timestamp := string(obj.Data)
		log.Printf("Received [%d:%d] %s", obj.Headers.GroupID, obj.Headers.ObjectID, timestamp)
	})

	log.Printf("Subscribed to %s...", ftn.String())
	log.Println("Press Ctrl+C to stop")

	// Wait for signal
	<-sigCh
	log.Println("Shutting down...")
}

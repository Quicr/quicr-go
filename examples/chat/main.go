// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

// Chat example - demonstrates subscribe namespace for multi-user chat.
//
// This example shows how to:
//   - Subscribe to a namespace to discover other chat participants
//   - Dynamically subscribe to tracks as they are announced
//   - Publish chat messages with proper grouping
//
// Usage:
//
//	chat -server localhost:4433 -session my-room
package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/quicr/qgo"
)

const (
	chatApp         = "chat"
	trackName       = "text"
	messagesPerGroup = 10
	defaultTTL      = 30000 // 30 seconds
	defaultPriority = 128
)

// ChatMessage represents a chat message payload.
type ChatMessage struct {
	From string `json:"from"`
	Name string `json:"name"`
	Text string `json:"text"`
	TS   int64  `json:"ts"`
}

// generateDID creates an ATProto-style DID from a username.
func generateDID(username string) string {
	// Generate random bytes for DID
	b := make([]byte, 20)
	rand.Read(b)
	encoded := base32.StdEncoding.EncodeToString(b)
	encoded = strings.ToLower(encoded[:24])
	return fmt.Sprintf("did:plc:%s", encoded)
}

func main() {
	server := flag.String("server", "localhost:4433", "Relay server address")
	transport := flag.String("transport", "quic", "Transport (quic or webtransport)")
	session := flag.String("session", "", "Chat session/room ID (required)")
	flag.Parse()

	if *session == "" {
		log.Fatal("Session ID required. Use -session <room-id>")
	}

	// Get username
	fmt.Print("Enter your username: ")
	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		log.Fatal("Username required")
	}

	// Generate DID
	userDID := generateDID(username)
	fmt.Printf("\nGenerating DID: %s\n", userDID)

	// Parse transport
	var transportType qgo.Transport
	switch *transport {
	case "quic":
		transportType = qgo.TransportQUIC
	case "webtransport", "wt":
		transportType = qgo.TransportWebTransport
	default:
		log.Fatalf("Unknown transport: %s", *transport)
	}

	// Create client
	cfg := qgo.ClientConfig{
		ConnectURI: *server,
		EndpointID: fmt.Sprintf("qgo-chat-%s", userDID),
		Transport:  transportType,
	}

	client, err := qgo.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	client.OnStatusChange(func(status qgo.ClientStatus) {
		log.Printf("Client status: %s", status)
	})

	// Connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Connecting to %s...", *server)
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	log.Println("Connected!")

	// Build namespaces
	commonNs := qgo.NewNamespace(chatApp, *session)
	userNs := qgo.NewNamespace(chatApp, *session, userDID)

	fmt.Println("\nNamespace tuples:")
	for i, entry := range commonNs.Entries() {
		fmt.Printf("  [%d] %s\n", i, string(entry))
	}

	fmt.Printf("\nSubscribe namespace: %s\n", commonNs.String())
	fmt.Printf("Publish track: %s/%s\n", userNs.String(), trackName)

	fmt.Println("\nTrack tuples:")
	for i, entry := range userNs.Entries() {
		fmt.Printf("  [%d] %s\n", i, string(entry))
	}
	fmt.Printf("\nTrack name: %s\n", trackName)

	// Create chat instance
	chat := &Chat{
		client:   client,
		userDID:  userDID,
		username: username,
		session:  *session,
		tracks:   make(map[string]*qgo.SubscribeTrackHandler),
	}

	// Set up the callback for when PUBLISH messages are received (SubNS flow)
	// This is called when another user starts publishing to a track matching our namespace subscription
	client.OnPublishReceived(func(ftn qgo.FullTrackName, trackAlias uint64, connHandle uint64, requestID uint64) {
		chat.onPublishReceived(ftn, trackAlias, connHandle, requestID)
	})

	// Subscribe to namespace
	if err := chat.subscribeNamespace(commonNs); err != nil {
		log.Fatalf("Failed to subscribe namespace: %v", err)
	}

	// Create publish track
	if err := chat.createPublishTrack(userNs); err != nil {
		log.Fatalf("Failed to create publish track: %v", err)
	}

	fmt.Println("\nEntered chat room. Type messages and press Enter to send. Ctrl+C to exit.")
	fmt.Println()

	// Signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Input handling
	inputCh := make(chan string)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputCh <- scanner.Text()
		}
		close(inputCh)
	}()

	// Main loop
	for {
		select {
		case <-sigCh:
			fmt.Println("\nLeaving chat...")
			chat.cleanup()
			return

		case text, ok := <-inputCh:
			if !ok {
				chat.cleanup()
				return
			}
			text = strings.TrimSpace(text)
			if text != "" {
				if err := chat.sendMessage(text); err != nil {
					log.Printf("Send error: %v", err)
				}
			}
		}
	}
}

// Chat manages the chat session.
type Chat struct {
	client   *qgo.Client
	userDID  string
	username string
	session  string

	nsHandler  *qgo.SubscribeNamespaceHandler
	pubHandler *qgo.PublishTrackHandler

	mu       sync.Mutex
	tracks   map[string]*qgo.SubscribeTrackHandler
	groupID  uint64
	objectID uint64
}

func (c *Chat) subscribeNamespace(ns qgo.Namespace) error {
	handler, err := qgo.NewSubscribeNamespaceHandler(ns)
	if err != nil {
		return err
	}
	c.nsHandler = handler

	handler.OnStatusChange(func(status qgo.SubscribeNamespaceStatus) {
		log.Printf("Namespace subscription: %s", status)
	})

	return c.client.SubscribeNamespace(handler)
}

func (c *Chat) onPublishReceived(ftn qgo.FullTrackName, trackAlias uint64, connHandle uint64, requestID uint64) {
	// Extract DID from namespace (3rd tuple: chat/session/did)
	entries := ftn.Namespace.Entries()
	if len(entries) < 3 {
		return
	}

	publisherDID := string(entries[2])

	// Don't subscribe to our own track
	if publisherDID == c.userDID {
		return
	}

	trackKey := ftn.String()

	c.mu.Lock()
	if _, exists := c.tracks[trackKey]; exists {
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()

	log.Printf("New participant: %s (track: %s, alias: %d)", publisherDID, trackKey, trackAlias)

	// Accept the PUBLISH — this sends PubOK and implicitly establishes subscription
	handler, err := c.client.ResolvePublish(connHandle, requestID, ftn, defaultPriority, qgo.PublishResolveOK)
	if err != nil {
		log.Printf("Failed to resolve publish from %s: %v", trackKey, err)
		return
	}

	handler.OnObjectReceived(func(obj qgo.Object) {
		c.onMessageReceived(publisherDID, obj)
	})

	c.mu.Lock()
	c.tracks[trackKey] = handler
	c.mu.Unlock()
}

func (c *Chat) onMessageReceived(senderDID string, obj qgo.Object) {
	var msg ChatMessage
	if err := json.Unmarshal(obj.Data, &msg); err != nil {
		log.Printf("Invalid message from %s: %v", senderDID, err)
		return
	}

	// Print message
	fmt.Printf("[%s] %s\n", msg.Name, msg.Text)
}

func (c *Chat) createPublishTrack(ns qgo.Namespace) error {
	// Create publish track (SubNS flow - no need for PublishNamespace/ANNOUNCE)
	ftn := qgo.FullTrackName{
		Namespace: ns,
		TrackName: qgo.NewTrackName(trackName),
	}

	cfg := qgo.PublishTrackConfig{
		FullTrackName: ftn,
		TrackMode:     qgo.TrackModeStream,
		Priority:      defaultPriority,
		TTL:           defaultTTL,
		UseAnnounce:   false, // SubNS flow - relay forwards PUBLISH to SubNS subscribers
	}

	handler, err := c.client.PublishTrack(cfg)
	if err != nil {
		return err
	}
	c.pubHandler = handler

	handler.OnStatusChange(func(status qgo.PublishStatus) {
		log.Printf("Publish status: %s", status)
	})

	// Initialize group ID from current time
	c.groupID = uint64(time.Now().Unix()) & 0xFFFFFFFF

	return nil
}

func (c *Chat) sendMessage(text string) error {
	if c.pubHandler == nil {
		return fmt.Errorf("not initialized")
	}

	if !c.pubHandler.CanPublish() {
		return fmt.Errorf("cannot publish (status: %s)", c.pubHandler.Status())
	}

	// Create message
	msg := ChatMessage{
		From: c.userDID,
		Name: c.username,
		Text: text,
		TS:   time.Now().UnixMilli(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Determine group/object IDs
	c.mu.Lock()
	groupID := c.groupID
	objectID := c.objectID
	c.objectID++
	if c.objectID >= messagesPerGroup {
		c.objectID = 0
		c.groupID++
	}
	c.mu.Unlock()

	headers := qgo.ObjectHeaders{
		GroupID:  groupID,
		ObjectID: objectID,
		Priority: defaultPriority,
	}

	status, err := c.pubHandler.PublishObject(headers, data)
	if err != nil {
		return err
	}

	if status.IsSuccess() {
		fmt.Printf("Published [%d:%d] %s\n", groupID, objectID, text)
	} else {
		return fmt.Errorf("publish failed: %s", status)
	}

	return nil
}

func (c *Chat) cleanup() {
	c.mu.Lock()
	for _, handler := range c.tracks {
		handler.Close()
	}
	c.tracks = nil
	c.mu.Unlock()

	if c.pubHandler != nil {
		c.pubHandler.Close()
	}
	if c.nsHandler != nil {
		c.nsHandler.Close()
	}
}

// SPDX-FileCopyrightText: Copyright (c) 2025 Cisco Systems
// SPDX-License-Identifier: BSD-2-Clause

// Package qgo provides Go bindings for libquicr, a Media over QUIC (MoQ) transport library.
//
// qgo enables Go applications to publish and subscribe to real-time media streams
// using the MoQ protocol over QUIC transport.
//
// # Quick Start
//
// To create a client and connect to a relay:
//
//	cfg := qgo.ClientConfig{
//		ConnectURI: "moqt://relay.example.com:4433",
//		EndpointID: "my-client",
//	}
//	client, err := qgo.NewClient(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	ctx := context.Background()
//	if err := client.Connect(ctx); err != nil {
//		log.Fatal(err)
//	}
//
// # Publishing
//
// To publish data to a track:
//
//	ns := qgo.NewNamespace("example", "room")
//	ftn := qgo.FullTrackName{
//		Namespace: ns,
//		TrackName: qgo.NewTrackName("video"),
//	}
//
//	handler, err := client.PublishTrack(qgo.PublishTrackConfig{
//		FullTrackName: ftn,
//		TrackMode:     qgo.TrackModeStream,
//		Priority:      128,
//		TTL:           5000,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	headers := qgo.ObjectHeaders{
//		GroupID:  0,
//		ObjectID: 1,
//		Priority: 128,
//	}
//	handler.PublishObject(headers, []byte("Hello, World!"))
//
// # Subscribing
//
// To subscribe to a track:
//
//	handler, err := client.SubscribeTrack(qgo.SubscribeTrackConfig{
//		FullTrackName: ftn,
//		Priority:      128,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	handler.OnObjectReceived(func(obj qgo.Object) {
//		fmt.Printf("Received: %s\n", string(obj.Data))
//	})
//
// # Thread Safety
//
// All public methods in this package are thread-safe and can be called
// concurrently from multiple goroutines. Callbacks are executed in
// separate goroutines to avoid blocking the underlying C transport.
package qgo

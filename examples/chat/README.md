# Chat Example

A multi-user chat application demonstrating the MoQ Subscribe Namespace (SubNS) flow.

## How It Works

1. Each user subscribes to namespace `chat/<session>` to discover other participants
2. Each user publishes to track `chat/<session>/<user-did>/text`
3. When a PUBLISH is received (via SubNS), the client subscribes to that track
4. Messages are received via the subscribed tracks

## Features

- SubNS flow for dynamic participant discovery
- ATProto-style DID generation for user identity
- Message grouping (10 messages per group)

## Building

```bash
# From the repo root
make examples

# Or build just the chat example
go build -o bin/chat ./examples/chat
```

## Running

### Prerequisites

You need a MoQ relay server running. Start one at `localhost:4433`.

### Start Chat

```bash
# Join a chat room
./bin/chat -server localhost:4433 -session my-room

# You'll be prompted for a username
Enter your username: alice
```

### Multiple Clients

Open multiple terminals and run the chat with the same session ID:

```bash
# Terminal 1
./bin/chat -session gaming-room

# Terminal 2
./bin/chat -session gaming-room
```

## Command Line Options

```
-server     Relay server address (default: localhost:4433)
-transport  Transport protocol: quic or webtransport (default: quic)
-session    Chat session/room ID (required)
```

## Namespace Structure

The chat uses a hierarchical namespace:

```
Subscribe namespace:  chat/<session-id>
Publish track:        chat/<session-id>/<user-did>/text
```

See [chat_design.md](chat_design.md) for detailed design documentation.

## Example Session

```
$ ./bin/chat -server localhost:4433 -session demo
Enter your username: alice

Generating DID: did:plc:abc123...

Namespace tuples:
  [0] chat
  [1] demo

Subscribe namespace: chat/demo
Publish track: chat/demo/did:plc:abc123.../text

Entered chat room. Type messages and press Enter to send. Ctrl+C to exit.

> Hello!
Published [1234567890:0] Hello!

[bob] Hey Alice!

> How's it going?
Published [1234567890:1] How's it going?

^C
Leaving chat...
```

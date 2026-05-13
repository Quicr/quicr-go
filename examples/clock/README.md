# Clock Example

A simple publish/subscribe example that broadcasts UTC timestamps.

## Usage

```bash
# Build
make examples

# Publish timestamps (once per second)
./bin/clock -mode publish -server localhost:4433

# Subscribe to timestamps (in another terminal)
./bin/clock -mode subscribe -server localhost:4433
```

## Options

```
-server       Relay server address (default: localhost:4433)
-transport    Transport: quic or webtransport (default: quic)
-namespace    Namespace path (default: clock/utc)
-track        Track name (default: time)
-mode         Mode: publish or subscribe
-priority     Priority 0-255 (default: 128)
-ttl          TTL in milliseconds (default: 5000)
-use_announce Use announce flow instead of direct publish
```

## Example Output

**Publisher:**
```
$ ./bin/clock -mode publish -server localhost:4433
Creating client with transport: QUIC, server: localhost:4433
Connecting to localhost:4433...
Connected!
Publishing UTC timestamps to clock/utc/time...
Published [0:1] 2026-05-12T21:30:00.123456789Z
Published [0:2] 2026-05-12T21:30:01.123456789Z
...
```

**Subscriber:**
```
$ ./bin/clock -mode subscribe -server localhost:4433
Connecting to localhost:4433...
Connected!
Subscribed to clock/utc/time...
Received [0:1] 2026-05-12T21:30:00.123456789Z
Received [0:2] 2026-05-12T21:30:01.123456789Z
...
```

<p align="center">
  <img src="logo.svg" alt="qgo logo" width="150">
</p>

# qgo - Go Bindings for libquicr

[![CI](https://github.com/Quicr/quicr-go/actions/workflows/ci.yml/badge.svg)](https://github.com/Quicr/quicr-go/actions/workflows/ci.yml)

Go bindings for [libquicr](https://github.com/Quicr/libquicr), a C++ implementation of Media over QUIC Transport (MoQT).

**Supports MoQT draft-16.**

## Installation

### Option 1: Using Pre-built Libraries (Recommended)

```bash
go get github.com/quicr/qgo@latest
go generate github.com/quicr/qgo
```

This downloads platform-specific pre-built libraries from the latest release.

### Option 2: Building from Source

```bash
git clone --recursive https://github.com/quicr/qgo.git
cd qgo
make shim
```

See [BUILD.md](BUILD.md) for detailed build instructions and requirements.

## Quick Start

```go
import "github.com/quicr/qgo"

func main() {
    // Create and connect client
    client, _ := qgo.NewClient(qgo.ClientConfig{
        ConnectURI: "moq://localhost:4433",
        EndpointID: "my-app",
    })
    defer client.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    client.Connect(ctx)

    // Publish objects
    pub, _ := client.PublishTrack(qgo.PublishTrackConfig{
        FullTrackName: qgo.FullTrackName{
            Namespace: qgo.ParseNamespace("app/channel"),
            TrackName: qgo.NewTrackName("data"),
        },
    })
    pub.PublishObject(qgo.ObjectHeaders{GroupID: 1, ObjectID: 1}, []byte("hello"))

    // Subscribe to objects
    sub, _ := client.SubscribeTrack(qgo.SubscribeTrackConfig{
        FullTrackName: qgo.FullTrackName{
            Namespace: qgo.ParseNamespace("app/channel"),
            TrackName: qgo.NewTrackName("data"),
        },
    })
    sub.OnObjectReceived(func(obj qgo.Object) {
        fmt.Println("Received:", string(obj.Data))
    })
}
```

## Examples

| Example | Description |
|---------|-------------|
| [clock](examples/clock/) | Simple publish/subscribe with timestamps |
| [chat](examples/chat/) | Multi-user chat using SubNS flow |

```bash
# Build examples
make examples

# Run clock publisher
./bin/clock -mode publish -server localhost:4433

# Run chat
./bin/chat -server localhost:4433 -session myroom
```

## Documentation

- [BUILD.md](BUILD.md) - Building from source
- [examples/](examples/) - Example applications with READMEs
- [pkg.go.dev](https://pkg.go.dev/github.com/quicr/qgo) - API documentation

## License

BSD-2-Clause. See [LICENSE](LICENSE).

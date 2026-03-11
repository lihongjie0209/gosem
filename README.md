# gosem

Go implementation of DLMS/COSEM protocol encoder/decoder for smart energy metering.

## Features

- **A-XDR Encoding/Decoding**: IEC 61334-6 A-XDR (Adjusted External Data Representation)
- **DLMS Protocol**: Full ACSE/COSEM services (AARQ/AARE, Get/Set/Action, notifications)
- **Multiple Transports**: HDLC, TCP Wrapper, Serial Port
- **Security**: AES-GCM ciphering support
- **Type-safe**: Fully typed interfaces with mock generation

## Installation

```bash
go get gitlab.com/circutor-library/gosem
```

## Quick Start

```go
package main

import (
    "time"
    "gitlab.com/circutor-library/gosem/pkg/dlms"
    "gitlab.com/circutor-library/gosem/pkg/dlmsclient"
    "gitlab.com/circutor-library/gosem/pkg/tcp"
    "gitlab.com/circutor-library/gosem/pkg/wrapper"
)

func main() {
    // Create transport stack (TCP + Wrapper)
    tcpTransport := tcp.New(4059, "192.168.1.34", 5*time.Second)
    wrapperTransport := wrapper.New(tcpTransport, 1, 1)
    
    // Configure client settings
    settings := dlms.Settings{
        Authentication: dlms.AuthenticationNone,
        // ... configure other settings
    }
    
    // Create DLMS client
    client := dlmsclient.New(settings, wrapperTransport, 10*time.Second, 0)
    
    // Read attribute
    attr := dlms.CreateAttributeDescriptor(3, "0-0:1.0.0.255", 2)
    result, err := client.Get(attr, nil)
    if err != nil {
        panic(err)
    }
}
```

## Architecture

### Core Packages

- **`pkg/axdr`**: Low-level A-XDR encoding/decoding with tagged data types
- **`pkg/dlms`**: DLMS protocol layer (ACSE, COSEM services, OBIS addressing)
- **`pkg/dlmsclient`**: High-level client with association management
- **`pkg/hdlc`**: HDLC framing (IEC 62056-46)
- **`pkg/wrapper`**: TCP wrapper protocol
- **`pkg/tcp`**: Raw TCP transport
- **`pkg/serialport`**: Serial port transport

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Generate mocks
go generate ./...
```

## OBIS Codes

Objects are addressed using OBIS codes (format `A-B:C.D.E.F`):
```go
attr := dlms.CreateAttributeDescriptor(classID, "0-0:1.0.0.255", attributeID)
```

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) file for details.

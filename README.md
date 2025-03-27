# NetLog

A lightweight network packet capture tool that monitors network traffic and provides real-time insights into your network activity.

## Features

- Real-time packet capture and monitoring
- Smart packet aggregation for large transfers
- Support for both text and JSON output formats
- Automatic detection of public IP traffic
- Dynamic window sizing based on transfer rates
- Timestamp tracking for all captured packets

## Installation

```bash
# Clone the repository
git clone https://github.com/highscaleco/netlog.git
cd netlog

# Build the project
go build -o netlog cmd/netlog/main.go
```

## Usage

### Basic Usage

```bash
# Run with default settings (captures on en1 interface)
sudo ./netlog

# Specify a different network interface
sudo ./netlog --interface eth0

# Output in JSON format
sudo ./netlog --format json
```

### Command Line Arguments

- `--interface`: Network interface to capture packets from (default: "en1")
- `--format`: Output format (default: "text")
  - `text`: Human-readable format
  - `json`: JSON-like string format

### Output Formats

#### Text Format
```
[2024-03-23 12:34:56] 192.168.1.100 => 8.8.8.8 TCP 443 500000000 bytes (50000 packets in 5.00s)
```

#### JSON Format
```json
{"timestamp":"2024-03-23 12:34:56","duration":"5.00s","source":"192.168.1.100","destination":"8.8.8.8","protocol":"TCP","port":"443","total_bytes":500000000,"packets":50000}
```

## Smart Aggregation

NetLog automatically adjusts its aggregation window based on transfer rates:

- ≤ 1 MB/s: 1-second window
- > 1 MB/s: 2-second window
- > 10 MB/s: 3-second window
- > 100 MB/s: 5-second window

This helps reduce output noise while maintaining meaningful progress updates for large transfers.

## Requirements

- Go 1.22 or later
- libpcap-dev (Linux) or libpcap (macOS)
- Root/sudo privileges for packet capture

### Linux Dependencies
```bash
sudo apt-get install libpcap-dev
```

### macOS Dependencies
```bash
brew install libpcap
```

## Development

### Project Structure

```
netlog/
├── cmd/
│   └── netlog/
│       └── main.go    # Main entry point
├── pkg/
│   ├── capture/      # Packet capture functionality
│   └── types/        # Shared types
└── README.md
```

### Building from Source

```bash
# Install dependencies
go mod download

# Build the project
go build -o netlog cmd/netlog/main.go
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Acknowledgments

- [gopacket](https://github.com/google/gopacket) for packet capture functionality
- [libpcap](https://www.tcpdump.org/) for the underlying packet capture library
# NetLog

A network traffic monitoring tool for Kubernetes clusters that captures and logs network packets, providing detailed information about network connections between pods.

## Features

- Captures network packets using libpcap
- Identifies Kubernetes pods by IP address
- Supports both TCP and UDP protocols
- Provides real-time logging of network connections
- JSON output format for easy parsing
- Prometheus metrics for monitoring and alerting

## Prerequisites

- Go 1.21 or later
- libpcap development files
- Kubernetes cluster access
- Redis (optional, for caching ovn-fip information)

## Installation

### Building from Source

1. Clone the repository:
```bash
git clone https://github.com/yourusername/netlog.git
cd netlog
```

2. Install dependencies:
```bash
go mod download
```

3. Build the binary:
```bash
go build -o netlog cmd/netlog/main.go
```

### Installing libpcap

On Ubuntu/Debian:
```bash
sudo apt-get install libpcap-dev
```

On macOS:
```bash
brew install libpcap
```

## Usage

### Basic Usage

Run NetLog with default settings:
```bash
sudo ./netlog
```

### Command Line Arguments

- `--interface`: Network interface to capture packets from (default: "eth0")
- `--redis-addr`: Redis server address (default: "localhost:6379")
- `--redis-password`: Redis password (optional)
- `--redis-db`: Redis database number (default: 0)
- `--json`: Enable JSON output format
- `--metrics-addr`: Address to expose Prometheus metrics (default: ":9090")

### Prometheus Metrics

NetLog exposes the following Prometheus metrics at the `/metrics` endpoint:

- `netlog_network_bytes_total`: Total bytes transferred
  - Labels: namespace, name, source, destination, protocol, port, direction
- `netlog_network_packets_total`: Total number of packets
  - Labels: namespace, name, source, destination, protocol, port, direction
- `netlog_network_connections_active`: Number of active connections
  - Labels: namespace, name, source, destination, protocol, port
- `netlog_network_connection_duration_seconds`: Duration of connections
  - Labels: namespace, name, source, destination, protocol, port

Example Prometheus queries:
```promql
# Total bytes transferred by namespace
sum(netlog_network_bytes_total) by (namespace)

# Active connections by pod
sum(netlog_network_connections_active) by (namespace, name)

# Average connection duration
rate(netlog_network_connection_duration_seconds_sum[5m]) / rate(netlog_network_connection_duration_seconds_count[5m])
```

## Output Format

### Text Output
```
[2024-02-14 12:34:56] namespace: default, name: nginx-7f9f9f9f9f, source: 10.244.1.2:80, destination: 10.244.2.3:443, protocol: TCP, bytes: 1234, packets: 10
```

### JSON Output
```json
{
  "timestamp": "2024-02-14T12:34:56Z",
  "namespace": "default",
  "name": "nginx-7f9f9f9f9f",
  "source": "10.244.1.2:80",
  "destination": "10.244.2.3:443",
  "protocol": "TCP",
  "bytes": 1234,
  "packets": 10
}
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
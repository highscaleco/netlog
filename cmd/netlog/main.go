package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/gopacket/pcap"
	"github.com/highscaleco/netlog/pkg/capture"
	"github.com/highscaleco/netlog/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var (
	// FormatFlag specifies the output format
	FormatFlag = "text"
	// InterfaceFlag specifies the network interface to capture from
	InterfaceFlag = "en1"
	// MetricsAddr specifies the address to expose metrics on
	MetricsAddr = ":9090"
)

var rootCmd = &cobra.Command{
	Use:   "netlog",
	Short: "A network packet capture tool",
	Long: `NetLog is a lightweight network packet capture tool that monitors network traffic
and provides real-time insights into your network activity.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize metrics
		metrics.Init()

		// Create capture instance
		capture := capture.NewCapture(
			InterfaceFlag,
			65536,             // bufferSize
			true,              // promiscuous
			pcap.BlockForever, // timeout
			"tcp or udp",      // filter
			65536,             // maxPacketSize
			10000,             // maxConnections
		)
		if capture == nil {
			return fmt.Errorf("failed to create capture instance")
		}

		// Start packet capture
		ctx := context.Background()
		if err := capture.Start(ctx); err != nil {
			return fmt.Errorf("failed to start capture: %v", err)
		}

		// Start metrics server
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			if err := http.ListenAndServe(MetricsAddr, nil); err != nil {
				fmt.Printf("Error starting metrics server: %v\n", err)
			}
		}()

		// Start metrics cleanup goroutine
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					metrics.CleanupMetrics()
				}
			}
		}()

		// Handle graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Process packets
		go func() {
			for packet := range capture.Packets() {
				var output string
				if FormatFlag == "json" {
					output = packet.JSONString()
				} else {
					output = packet.String()
				}
				if output != "" {
					fmt.Println(output)
				}

				// Update Prometheus metrics
				if packet.Namespace != "" {
					metrics.UpdateMetrics(
						packet.Namespace,
						packet.Name,
						packet.Source,
						packet.Destination,
						packet.Protocol,
						packet.Port,
						packet.Direction,
						packet.TotalBytes,
						packet.Packets,
						packet.EndTime.Sub(packet.StartTime).Seconds(),
					)
				}
			}
		}()

		// Wait for shutdown signal
		<-sigChan
		capture.Stop()

		return nil
	},
}

func init() {
	rootCmd.Flags().StringVarP(&FormatFlag, "format", "f", "text", "Output format (text or json)")
	rootCmd.Flags().StringVarP(&InterfaceFlag, "interface", "i", "eth0", "Network interface to capture from")
	rootCmd.Flags().StringVarP(&MetricsAddr, "metrics-addr", "m", ":9090", "Address to expose metrics on")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/highscaleco/netlog/pkg/capture"
	"github.com/spf13/cobra"
)

var (
	// FormatFlag specifies the output format
	FormatFlag = "text"
	// InterfaceFlag specifies the network interface to capture from
	InterfaceFlag = "en1"
)

var rootCmd = &cobra.Command{
	Use:   "netlog",
	Short: "A network packet capture tool",
	Long: `NetLog is a lightweight network packet capture tool that monitors network traffic
and provides real-time insights into your network activity.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create capture instance
		capture := capture.NewCapture(InterfaceFlag)
		if capture == nil {
			return fmt.Errorf("failed to create capture instance")
		}

		// Start packet capture
		packets, err := capture.Start()
		if err != nil {
			return fmt.Errorf("failed to start capture: %v", err)
		}

		// Handle graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Process packets
		go func() {
			for packet := range packets {
				var output string
				if FormatFlag == "json" {
					output = packet.JSONString()
				} else {
					output = packet.String()
				}
				if output != "" {
					fmt.Println(output)
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
	rootCmd.Flags().StringVarP(&InterfaceFlag, "interface", "i", "en1", "Network interface to capture from")
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

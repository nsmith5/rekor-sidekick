package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCLI() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rekor-sidekick",
		Short: "Transparency log monitoring and alerting",
		Run:   runCLI,
	}
	cmd.Flags().String("config", "/etc/rekor-sidekick/config.yaml", "Path to configuration file")
	return cmd
}

func runCLI(cmd *cobra.Command, args []string) {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		fmt.Println("Failed to bind command line flags to viper:", err)
		os.Exit(1)
	}

	// Environment variable setup
	viper.SetEnvPrefix(`REKOR_SIDEKICK`)
	viper.AutomaticEnv()

	// Load config file
	{
		f, err := os.Open(viper.GetString("config"))
		if err != nil {
			fmt.Println("Failed to open config file:", err)
			os.Exit(1)
		}
		defer f.Close()

		viper.SetConfigType("yaml")
		if err := viper.ReadConfig(f); err != nil {
			fmt.Println("Failed to parse config:", err)
			os.Exit(1)
		}
	}

	var c config
	if err := viper.Unmarshal(&c); err != nil {
		fmt.Println("Failed to load configuration:", err)
		os.Exit(1)
	}

	a, err := newAgent(c)
	if err != nil {
		fmt.Println("Failed to initialize agent:", err)
		os.Exit(1)
	}

	// This run group manages all the concurrent processes
	// this command runs. That is the agent and a signal handler
	// at the moment.
	var g run.Group

	// Agent process
	g.Add(a.run, func(error) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		a.shutdown(ctx)
	})

	// Signal handler process
	{
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-ctx.Done():
				return ctx.Err()
			}
		}, func(error) {
			cancel()
		})
	}

	// Launch!
	if err = g.Run(); err != nil {
		log.Println(err)
	}
}

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JanKrl/homelab-version-agent/internal/checker"
	"github.com/JanKrl/homelab-version-agent/internal/config"
	"github.com/JanKrl/homelab-version-agent/internal/publisher"
)

var version = "v0.1.0-dev"

func main() {
	var configDir = flag.String("config", "/etc/homelab-agent/config.yaml", "Configuration file location")
	var showVersion = flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("homelab-version-agent %s\n", version)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("Starting homelab version agent", "config", *configDir)

	cfg, err := config.GetConfig(*configDir)
	if err != nil {
		slog.Error("Unable to get config", "error", err)
		os.Exit(1)
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	var checkers []checker.Checker
	for _, target := range cfg.Targets.AptPackages {
		checkers = append(checkers, checker.NewAptChecker(target))
	}
	for _, target := range cfg.Targets.GitHubRepos {
		checkers = append(checkers, checker.NewGitHubChecker(target.Owner, target.Name, target.Cmd, httpClient))
	}

	pub, err := publisher.NewMqttPublisher(cfg.Mqtt, cfg.Host)
	if err != nil {
		slog.Error("Unable to initialize MQTT publisher", "error", err)
		os.Exit(1)
	}

	slog.Info("Initialized, checking...")
	runCheck(ctx, checkers, pub)

	ticker := time.NewTicker(cfg.RunInterval)
	defer ticker.Stop()

	slog.Info("Done. Entering loop mode", "interval", cfg.RunInterval.String())
	for {
		select {
		case <-ctx.Done():
			slog.Info("Shutting down. Bye!")
			return
		case <-ticker.C:
			slog.Info("Running scheduled check...")
			runCheck(ctx, checkers, pub)
			slog.Info("Done!")
		}
	}
}

func runCheck(ctx context.Context, checkers []checker.Checker, pub publisher.Publisher) {
	var results []checker.Result
	for _, checker := range checkers {
		result, err := checker.Check(ctx)
		if err != nil {
			slog.Warn("Unable to check", "checker", checker)
			continue
		}
		results = append(results, result)
	}

	if err := pub.Publish(ctx, results); err != nil {
		slog.Error("Failed to publish results", "error", err)
	}
}

// Package main provides the command-line entrypoint for Digital Exhaust Cleaner.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"digital-exhaust-cleaner/internal/analyzer"
	"digital-exhaust-cleaner/internal/config"
	"digital-exhaust-cleaner/internal/logging"
	"digital-exhaust-cleaner/internal/ui"
	"go.uber.org/zap"
)

func main() {
	if err := run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return errors.New("expected command: scan or serve")
	}

	switch args[1] {
	case "scan":
		return runScan(ctx, args[2:])
	case "serve":
		return runServe(ctx, args[2:])
	default:
		return fmt.Errorf("unknown command %q", args[1])
	}
}

func runScan(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	configPath := fs.String("config", "configs/default.yaml", "path to YAML configuration")
	rootPath := fs.String("path", ".", "directory to scan")
	reportPath := fs.String("report", "reports/scan.html", "path to write the local HTML report")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}

	logger, err := logging.New(cfg.Logging)
	if err != nil {
		return err
	}
	defer logger.Sync()

	result, err := analyzer.New(cfg, logger).Analyze(ctx, *rootPath)
	if err != nil {
		return err
	}
	if err := ui.WriteReport(*reportPath, result); err != nil {
		return err
	}

	logger.Info(
		"scan complete",
		zap.String("root", result.Root),
		zap.Int64("files_scanned", result.FilesScanned),
		zap.Int("duplicate_groups", len(result.DuplicateGroups)),
		zap.Int("similar_groups", len(result.SimilarGroups)),
		zap.Int("recommendations", len(result.Recommendations)),
		zap.String("report", *reportPath),
	)

	for _, rec := range result.Recommendations {
		fmt.Printf("%s\t%.2f\t%s\n", rec.Category, rec.Score, rec.Explanation)
	}

	return nil
}

func runServe(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	configPath := fs.String("config", "configs/default.yaml", "path to YAML configuration")
	rootPath := fs.String("path", ".", "directory to scan")
	addr := fs.String("addr", "127.0.0.1:8787", "loopback address for the local UI")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}

	logger, err := logging.New(cfg.Logging)
	if err != nil {
		return err
	}
	defer logger.Sync()

	result, err := analyzer.New(cfg, logger).Analyze(ctx, *rootPath)
	if err != nil {
		return err
	}

	root, err := filepath.Abs(*rootPath)
	if err != nil {
		return fmt.Errorf("resolve root: %w", err)
	}

	logger.Info(
		"interactive UI ready",
		zap.String("url", "http://"+*addr),
		zap.String("root", root),
		zap.Int64("files_scanned", result.FilesScanned),
		zap.Int("recommendations", len(result.Recommendations)),
	)
	fmt.Printf("Open http://%s in your browser. Press Ctrl+C to stop.\n", *addr)

	return ui.Serve(ctx, ui.ServerConfig{
		Addr:          *addr,
		Root:          root,
		QuarantineDir: cfg.App.QuarantineDir,
		Result:        result,
		ScanFunc: func(scanCtx context.Context, targetPath string) (analyzer.Result, error) {
			return analyzer.New(cfg, logger).Analyze(scanCtx, targetPath)
		},
	})
}

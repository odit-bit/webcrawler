package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/odit-bit/indexstore"
	"github.com/odit-bit/linkstore"
	"github.com/odit-bit/webcrawler/crawler"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/urfave/cli/v2"
)

func main() {

	var (
	// linkstoreAddress  string
	// indexstoreAddress string
	// otelExporterHost  string
	)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	app := cli.NewApp()
	app.Name = "Link-Crawler"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "graph-api",
			Usage:   "address of graph grpc server",
			EnvVars: []string{"LINKSTORE_SERVER_ADDRESS"},

			Action: func(ctx *cli.Context, s string) error {
				if s == "" {
					return fmt.Errorf("graph grpc server address not set")
				}
				return nil
			},
		},

		&cli.StringFlag{
			Name:    "index-api",
			Usage:   "addres of index grpc server",
			EnvVars: []string{"INDEXSTORE_SERVER_ADDRESS"},

			Action: func(ctx *cli.Context, s string) error {
				if s == "" {
					return fmt.Errorf("index grpc server address not set")
				}
				return nil
			},
		},

		&cli.StringFlag{
			Name:    "tracer-host",
			Usage:   "address exporter host",
			EnvVars: []string{"OTEL_EXPORTER_HOST"},
		},

		&cli.DurationFlag{
			Name:    "update-interval",
			Value:   1 * time.Minute,
			Usage:   "duration of crawler to wake again",
			EnvVars: []string{"CRAWL_UPDATE_INTERVAL"},
		},

		&cli.DurationFlag{
			Name:    "update-threshold",
			Value:   7 * 24 * time.Hour,
			Usage:   "minimum time before link to crawled (update)",
			EnvVars: []string{"CRAWL_UPDATE_THRESHOLD"},
		},
	}

	app.Action = action()

	if err := app.Run(os.Args); err != nil {
		logger.Error(err.Error())
		return
	}

	logger.Warn("exit crawler")
}

func action() cli.ActionFunc {

	return func(cliCtx *cli.Context) error {
		var wg sync.WaitGroup

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		exporter, err := newGrpcExporter(ctx, cliCtx.String("tracer-host"))
		if err != nil {
			return err
		}

		shutdownFunc, err := setupOTelSDK(ctx, cliCtx.App.Name, cliCtx.App.Version, exporter)
		if err != nil {
			return err
		}
		// Handle shutdown properly so nothing leaks.
		defer func() {
			err = errors.Join(err, shutdownFunc(context.Background()))
		}()

		//setup server connection
		graphAPI, err := linkstore.ConnectGraph(cliCtx.String("graph-api"))
		if err != nil {
			return fmt.Errorf("graph-api: %v", err)
		}

		indexAPI, err := indexstore.ConnectIndex(cliCtx.String("index-api"))
		if err != nil {
			return fmt.Errorf("index-api: %v", err)
		}

		//setup crawler
		conf := crawler.Config{
			Interval:        cliCtx.Duration("update-interval"),
			ReCrawlTreshold: cliCtx.Duration("update-threshold"),
			Tracer:          otel.Tracer("crawler"),
		}
		cr := crawler.NewConfig(graphAPI, indexAPI, &conf)

		//setup gracefull
		sigC := make(chan os.Signal, 1)
		signal.Notify(sigC, syscall.SIGINT)

		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-sigC:
				cancel()
			case <-ctx.Done():
			}
		}()

		err = cr.Run(ctx)
		wg.Wait()
		return err
	}
}

func newGrpcExporter(ctx context.Context, host string) (sdktrace.SpanExporter, error) {

	nCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var opts []otlptracegrpc.Option

	opts = append(opts, otlptracegrpc.WithInsecure())
	opts = append(opts, otlptracegrpc.WithEndpoint(host))

	exp, err := otlptracegrpc.New(nCtx, opts...)
	if err != nil {
		return nil, err
	}
	return exp, nil

}

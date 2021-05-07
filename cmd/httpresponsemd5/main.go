package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpresponsemd5 "github.com/nicolasparada/go-http-response-md5"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := log.Default()
	err := run(ctx, logger, os.Args[1:])
	if err != nil {
		logger.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *log.Logger, args []string) error {
	var (
		concurrency uint64
		timeout     time.Duration
	)

	fs := flag.NewFlagSet("httpresponsemd5", flag.ExitOnError)
	fs.Uint64Var(&concurrency, "parallel", 10, "Max number of concurrent request to perform")
	fs.DurationVar(&timeout, "timeout", 0, "HTTP requests timeout duration")
	fs.Usage = func() {
		fmt.Println("httpresponsemd5 [flags] url [url2 [url3] ...]\nFlags:")
		fs.PrintDefaults()
	}

	err := fs.Parse(args)
	if err != nil {
		return fmt.Errorf("could not parse flags: %w", err)
	}

	go func() {
		<-ctx.Done()
		fmt.Println()
	}()

	urls := fs.Args()
	svc := &httpresponsemd5.Service{
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Concurrency: concurrency,
	}
	rr := svc.CollectMD5Hashes(ctx, urls...)
	for result := range rr {
		if result.Err != nil {
			logger.Printf("failed to compute md5 hash out of %q: %v\n", result.RawURL, result.Err)
			continue
		}

		fmt.Printf("%s %s\n", result.RawURL, result.MD5Hash)
	}

	return nil
}

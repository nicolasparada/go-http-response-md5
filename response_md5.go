package httpresponsemd5

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var ErrInvalidURL = errors.New("invalid URL")

type Service struct {
	HTTPClient  *http.Client
	Concurrency uint64
}

type Result struct {
	RawURL  string
	MD5Hash string
	Err     error
}

func (svc *Service) CollectMD5Hashes(ctx context.Context, rawurls ...string) <-chan Result {
	results := make(chan Result)

	go func() {
		// g wait group so we can defer closing of channels.
		var g sync.WaitGroup
		g.Add(len(rawurls))

		defer close(results)

		var sem chan struct{}
		if svc.Concurrency != 0 {
			sem = make(chan struct{}, svc.Concurrency)
			defer close(sem)
		}

		for _, rawurl := range rawurls {
			if svc.Concurrency != 0 {
				sem <- struct{}{}
			}

			go func(rawurl string) {
				if svc.Concurrency != 0 {
					defer func() {
						<-sem
					}()
				}
				defer g.Done()

				md5Hash, err := svc.md5(ctx, rawurl)
				results <- Result{
					RawURL:  rawurl,
					MD5Hash: md5Hash,
					Err:     err,
				}
			}(rawurl)
		}

		g.Wait()
	}()

	return results
}

func (svc *Service) md5(ctx context.Context, rawurl string) (string, error) {
	rawurl = strings.TrimSpace(rawurl)
	if rawurl == "" {
		return "", ErrInvalidURL
	}

	u, err := url.Parse(rawurl)
	if err != nil {
		return "", ErrInvalidURL
	}

	if u.Scheme == "" {
		u.Scheme = "http"
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", ErrInvalidURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("coul not create http request: %w", err)
	}

	resp, err := svc.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not do http request: %w", err)
	}

	defer resp.Body.Close()

	h := md5.New()
	_, err = io.Copy(h, resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not generate md5 hash out of response body: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

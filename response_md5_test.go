package httpresponsemd5

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestService_md5(t *testing.T) {
	tt := []struct {
		name    string
		rawurl  string
		wantErr error
	}{
		{
			name:    "empty_url",
			wantErr: ErrInvalidURL,
		},
		{
			name:    "blank_url",
			rawurl:  "         ",
			wantErr: ErrInvalidURL,
		},
		{
			name: "invalid_url",
			// taken from https://golang.org/src/net/url/url_test.go
			rawurl:  "[fe80::1%en0]",
			wantErr: ErrInvalidURL,
		},
		{
			name:    "unsuported_scheme",
			rawurl:  "nope://nope",
			wantErr: ErrInvalidURL,
		},
		{
			name:   "default_scheme",
			rawurl: "example.org",
		},
		{
			name:   "ok",
			rawurl: "https://example.org",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			svc := &Service{
				HTTPClient: http.DefaultClient,
			}

			_, err := svc.md5(context.Background(), tc.rawurl)
			if err != tc.wantErr {
				t.Errorf("Service.md5() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
		})
	}
}

func TestService_CollectMD5Hashes(t *testing.T) {
	data := []byte("test")
	h := md5.New()
	_, err := h.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	md5Hash := fmt.Sprintf("%x", h.Sum(nil))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(data)
		if err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	tt := []struct {
		name        string
		concurrency uint64
		rawurls     []string
		test        func(t *testing.T, got []Result)
	}{
		{
			name:    "concurrency_cap_disabled",
			rawurls: []string{"", ""},
			test: func(t *testing.T, got []Result) {
				if want, got := 2, len(got); want != got {
					t.Errorf("result length: want %d; got %d", want, got)
				}

				if want := []Result{{Err: ErrInvalidURL}, {Err: ErrInvalidURL}}; !reflect.DeepEqual(want, got) {
					t.Errorf("results: want %+v; got %+v", want, got)
				}
			},
		},
		{
			name:        "concurrency_cap_one",
			concurrency: 1,
			rawurls:     []string{"", ""},
			test: func(t *testing.T, got []Result) {
				if want, got := 2, len(got); want != got {
					t.Errorf("result length: want %d; got %d", want, got)
				}

				if want := []Result{{Err: ErrInvalidURL}, {Err: ErrInvalidURL}}; !reflect.DeepEqual(want, got) {
					t.Errorf("results: want %+v; got %+v", want, got)
				}
			},
		},
		{
			name:        "concurrency_cap_three",
			concurrency: 3,
			rawurls:     []string{"", "", ""},
			test: func(t *testing.T, got []Result) {
				if want, got := 3, len(got); want != got {
					t.Errorf("result length: want %d; got %d", want, got)
				}

				if want := []Result{{Err: ErrInvalidURL}, {Err: ErrInvalidURL}, {Err: ErrInvalidURL}}; !reflect.DeepEqual(want, got) {
					t.Errorf("results: want %+v; got %+v", want, got)
				}
			},
		},
		{
			name:    "ok",
			rawurls: []string{srv.URL},
			test: func(t *testing.T, got []Result) {
				if want, got := 1, len(got); want != got {
					t.Errorf("result length: want %d; got %d", want, got)
				}

				if want := []Result{{RawURL: srv.URL, MD5Hash: md5Hash}}; !reflect.DeepEqual(want, got) {
					t.Errorf("results: want %+v; got %+v", want, got)
				}
			},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			svc := &Service{
				HTTPClient:  http.DefaultClient,
				Concurrency: tc.concurrency,
			}
			results := svc.CollectMD5Hashes(context.Background(), tc.rawurls...)
			var got []Result
			for result := range results {
				got = append(got, result)
			}

			tc.test(t, got)
		})
	}
}

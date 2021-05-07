# HTTP Response MD5

This generates MD5 hashes out of the response bodies of HTTP requests of the given URLs.

## Build

You need to have [Golang](https://golang.org/) installed.

```bash
$ go build ./cmd/httpresponsemd5
```

## Usage

```bash
$ ./httpresponsemd5 -h
httpresponsemd5 [flags] url [url2 [url3] ...]
Flags:
  -parallel uint
        Max number of concurrent request to perform (default 10)
  -timeout duration
        HTTP requests timeout duration
```

Example:
```bash
$ ./httpresponsemd5 https://example.org
https://example.org 84238dfc8092e5d9c0dac8ef93371a07
```

```bash
$ ./httpresponsemd5 https://example.org https://google.com
https://google.com 8a7d1076a1a2b3cc047e35789a786bbb
https://example.org 84238dfc8092e5d9c0dac8ef93371a07
```

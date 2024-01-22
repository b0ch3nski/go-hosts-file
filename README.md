# go-hosts-file
[![license](https://img.shields.io/github/license/b0ch3nski/go-hosts-file)](LICENSE)
[![release](https://img.shields.io/github/v/release/b0ch3nski/go-hosts-file)](https://github.com/b0ch3nski/go-hosts-file/releases)
[![go.dev](https://pkg.go.dev/badge/github.com/b0ch3nski/go-hosts-file)](https://pkg.go.dev/github.com/b0ch3nski/go-hosts-file)
[![goreportcard](https://goreportcard.com/badge/github.com/b0ch3nski/go-hosts-file)](https://goreportcard.com/report/github.com/b0ch3nski/go-hosts-file)
[![issues](https://img.shields.io/github/issues/b0ch3nski/go-hosts-file)](https://github.com/b0ch3nski/go-hosts-file/issues)
[![sourcegraph](https://sourcegraph.com/github.com/b0ch3nski/go-hosts-file/-/badge.svg)](https://sourcegraph.com/github.com/b0ch3nski/go-hosts-file)

Yet another Golang library for hosts file manipulation. Written for my own purposes but can be easily adopted for
multiple other use cases like merging hosts files, removing duplicates, validating, reverse lookups etc.

## install

```
go get github.com/b0ch3nski/go-hosts-file
```

## example

```go
hostsFile, _ := os.Open("/etc/hosts")

h := hosts.New()
h.Read(hostsFile)
h.Add(netip.AddrFrom4([4]byte{8, 8, 8, 8}), "google.com")

fmt.Print(&h)
fmt.Println(h.GetIP("localhost"))
fmt.Println(h.GetAlias(netip.AddrFrom4([4]byte{127, 0, 0, 1})))
```

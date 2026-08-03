package main

import "fmt"

// These variables are populated at build time using -ldflags
var (
	Version    = "unknown"
	CommitHash = "unknown"
	BuildDate  = "unknown"
)

type ProductVersion struct {
	Version    string
	CommitHash string
	BuildDate  string
}

func GetProductVersion() string {
	msg := fmt.Sprintf("v%s (%s) compiled on %s", Version, CommitHash, BuildDate)
	return msg
}

package main

import "fmt"

// These variables are populated at build time using -ldflags
var (
	Version    = "0.0.0"
	CommitHash = "0000000"
	BuildDate  = "1900-01-01"
)

type ProductVersion struct {
	Version    string
	CommitHash string
	BuildDate  string
}

func GetProductVersion() string {
	msg := fmt.Sprintf("%s (%s) compiled on %s", Version, CommitHash, BuildDate)
	return msg
}

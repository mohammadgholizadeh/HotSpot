//go:build ignore

package main

import (
    "log"

    // This file is ignored by builds. The active entrypoint is cmd/cli/main.go
    "github.com/mohammadgholizadeh/hotspot/cmd/cli"
)

func main() {
    if err := cli.Execute(); err != nil {
        log.Fatal(err)
    }
}

// Package config contains configuration.
package config

import "time"

// Config contains configuration.
type Config struct {
	SleepDuration time.Duration
	Board         []string
	MaxPrintDepth int
	EnablePrint   bool
	EnableShow    bool
	PrintDepth    bool
}

// Package config contains configuration.
package config

import "time"

// Config contains configuration.
type Config struct {
	SleepDuration   time.Duration
	Board           []string
	MaxPrintDepth   int
	EnableShow      bool
	PrintDepth      bool
	EnablePromotion bool
	EnableDrop      bool
}

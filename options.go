package mtgjsonsdk

import (
	"time"

	"github.com/mtgjson/mtgjson-sdk-go/db"
)

// Option configures the SDK.
type Option func(*db.Config)

// WithCacheDir sets the directory for cached parquet and JSON files.
func WithCacheDir(dir string) Option {
	return func(c *db.Config) {
		c.CacheDir = dir
	}
}

// WithOffline disables network requests; only cached data is used.
func WithOffline(offline bool) Option {
	return func(c *db.Config) {
		c.Offline = offline
	}
}

// WithTimeout sets the HTTP request timeout for downloads.
func WithTimeout(d time.Duration) Option {
	return func(c *db.Config) {
		c.Timeout = d
	}
}

// WithProgress sets a callback for download progress reporting.
func WithProgress(fn db.ProgressFunc) Option {
	return func(c *db.Config) {
		c.OnProgress = fn
	}
}

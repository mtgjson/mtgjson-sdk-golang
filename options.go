package mtgjson

import "time"

// ProgressFunc is called during file downloads to report progress.
// filename is the CDN file name, downloaded is bytes received so far,
// total is the total expected bytes (0 if unknown).
type ProgressFunc func(filename string, downloaded, total int64)

// Option configures the SDK.
type Option func(*sdkConfig)

type sdkConfig struct {
	cacheDir   string
	offline    bool
	timeout    time.Duration
	onProgress ProgressFunc
}

func defaultConfig() *sdkConfig {
	return &sdkConfig{
		cacheDir: defaultCacheDir(),
		offline:  false,
		timeout:  120 * time.Second,
	}
}

// WithCacheDir sets the directory for cached parquet and JSON files.
func WithCacheDir(dir string) Option {
	return func(c *sdkConfig) {
		c.cacheDir = dir
	}
}

// WithOffline disables network requests; only cached data is used.
func WithOffline(offline bool) Option {
	return func(c *sdkConfig) {
		c.offline = offline
	}
}

// WithTimeout sets the HTTP request timeout for downloads.
func WithTimeout(d time.Duration) Option {
	return func(c *sdkConfig) {
		c.timeout = d
	}
}

// WithProgress sets a callback for download progress reporting.
func WithProgress(fn ProgressFunc) Option {
	return func(c *sdkConfig) {
		c.onProgress = fn
	}
}

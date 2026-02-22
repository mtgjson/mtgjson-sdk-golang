package db

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// CacheManager downloads and caches MTGJSON data files from the CDN.
// It checks Meta.json for version changes and re-downloads when stale.
type CacheManager struct {
	CacheDir   string
	Offline    bool
	Timeout    int64 // seconds
	onProgress ProgressFunc

	client        *http.Client
	clientOnce    sync.Once
	remoteVer     string
	remoteVerOnce sync.Once
	mu            sync.Mutex
}

// NewCacheManager creates a CacheManager from the given Config.
func NewCacheManager(cfg *Config) (*CacheManager, error) {
	cm := &CacheManager{
		CacheDir:   cfg.CacheDir,
		Offline:    cfg.Offline,
		Timeout:    int64(cfg.Timeout.Seconds()),
		onProgress: cfg.OnProgress,
	}
	if err := os.MkdirAll(cm.CacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("mtgjson: create cache dir: %w", err)
	}
	return cm, nil
}

func (m *CacheManager) httpClient() *http.Client {
	m.clientOnce.Do(func() {
		m.client = &http.Client{
			Timeout: 0, // we handle timeouts per-request via context
		}
	})
	return m.client
}

// Close releases the HTTP client resources.
func (m *CacheManager) Close() {
	if m.client != nil {
		m.client.CloseIdleConnections()
	}
}

func (m *CacheManager) localVersion() string {
	data, err := os.ReadFile(filepath.Join(m.CacheDir, "version.txt"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func (m *CacheManager) saveVersion(version string) {
	_ = os.WriteFile(filepath.Join(m.CacheDir, "version.txt"), []byte(version), 0o644)
}

// RemoteVersion fetches the current MTGJSON version from Meta.json on the CDN.
// Returns empty string if offline or unreachable.
func (m *CacheManager) RemoteVersion(ctx context.Context) string {
	if m.remoteVer != "" {
		return m.remoteVer
	}
	if m.Offline {
		return ""
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, MetaURL, nil)
	if err != nil {
		return ""
	}
	resp, err := m.httpClient().Do(req)
	if err != nil {
		slog.Warn("Failed to fetch MTGJSON version from CDN", "error", err)
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Warn("MTGJSON CDN returned non-200", "status", resp.StatusCode)
		return ""
	}
	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return ""
	}
	// Try data.version, then meta.version
	if d, ok := data["data"].(map[string]any); ok {
		if v, ok := d["version"].(string); ok && v != "" {
			m.remoteVer = v
			return v
		}
	}
	if d, ok := data["meta"].(map[string]any); ok {
		if v, ok := d["version"].(string); ok && v != "" {
			m.remoteVer = v
			return v
		}
	}
	return ""
}

// IsStale checks if local cache is out of date compared to the CDN.
func (m *CacheManager) IsStale(ctx context.Context) bool {
	local := m.localVersion()
	remote := m.RemoteVersion(ctx)
	if remote == "" {
		return false // can't check, assume fresh
	}
	if local == "" {
		return true // no local version
	}
	return local != remote
}

func (m *CacheManager) downloadFile(ctx context.Context, filename string, dest string) error {
	url := CDNBase + "/" + filename
	slog.Info("Downloading", "url", url)

	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	tmpDest := dest + ".tmp"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := m.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", filename, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: HTTP %d", filename, resp.StatusCode)
	}

	total := resp.ContentLength
	f, err := os.Create(tmpDest)
	if err != nil {
		return err
	}

	var downloaded int64
	buf := make([]byte, 65536)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := f.Write(buf[:n]); wErr != nil {
				f.Close()
				os.Remove(tmpDest)
				return wErr
			}
			downloaded += int64(n)
			if m.onProgress != nil {
				m.onProgress(filename, downloaded, total)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			f.Close()
			os.Remove(tmpDest)
			return fmt.Errorf("download %s: %w", filename, readErr)
		}
	}
	f.Close()

	if err := os.Rename(tmpDest, dest); err != nil {
		os.Remove(tmpDest)
		return err
	}
	return nil
}

// EnsureParquet ensures a parquet file is cached locally, downloading if needed.
func (m *CacheManager) EnsureParquet(ctx context.Context, viewName string) (string, error) {
	filename, ok := ParquetFiles[viewName]
	if !ok {
		return "", fmt.Errorf("mtgjson: unknown parquet view %q", viewName)
	}
	localPath := filepath.Join(m.CacheDir, filename)

	m.mu.Lock()
	defer m.mu.Unlock()

	exists := fileExists(localPath)
	if !exists || m.IsStale(ctx) {
		if m.Offline {
			if exists {
				return localPath, nil
			}
			return "", fmt.Errorf("mtgjson: parquet file %s not cached and offline mode is enabled", filename)
		}
		if err := m.downloadFile(ctx, filename, localPath); err != nil {
			return "", err
		}
		if v := m.RemoteVersion(ctx); v != "" {
			m.saveVersion(v)
		}
	}
	return localPath, nil
}

// EnsureJSON ensures a JSON file is cached locally, downloading if needed.
func (m *CacheManager) EnsureJSON(ctx context.Context, name string) (string, error) {
	filename, ok := JSONFiles[name]
	if !ok {
		return "", fmt.Errorf("mtgjson: unknown JSON file %q", name)
	}
	localPath := filepath.Join(m.CacheDir, filename)

	m.mu.Lock()
	defer m.mu.Unlock()

	exists := fileExists(localPath)
	if !exists || m.IsStale(ctx) {
		if m.Offline {
			if exists {
				return localPath, nil
			}
			return "", fmt.Errorf("mtgjson: JSON file %s not cached and offline mode is enabled", filename)
		}
		if err := m.downloadFile(ctx, filename, localPath); err != nil {
			return "", err
		}
		if v := m.RemoteVersion(ctx); v != "" {
			m.saveVersion(v)
		}
	}
	return localPath, nil
}

// LoadJSON loads and parses a JSON file (handles .gz transparently).
func (m *CacheManager) LoadJSON(ctx context.Context, name string) (map[string]any, error) {
	path, err := m.EnsureJSON(ctx, name)
	if err != nil {
		return nil, err
	}
	return readJSONFile(path)
}

// Clear removes all cached files and recreates the cache directory.
func (m *CacheManager) Clear() error {
	if err := os.RemoveAll(m.CacheDir); err != nil {
		return err
	}
	return os.MkdirAll(m.CacheDir, 0o755)
}

// ResetRemoteVersion clears the cached remote version so it's re-fetched.
func (m *CacheManager) ResetRemoteVersion() {
	m.remoteVer = ""
	m.remoteVerOnce = sync.Once{}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func readJSONFile(path string) (map[string]any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var reader io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		gr, err := gzip.NewReader(f)
		if err != nil {
			// Corrupt gz file — remove it
			f.Close()
			os.Remove(path)
			return nil, fmt.Errorf("mtgjson: corrupt cache file %s: %w", filepath.Base(path), err)
		}
		defer gr.Close()
		reader = gr
	}

	var result map[string]any
	if err := json.NewDecoder(reader).Decode(&result); err != nil {
		// Corrupt JSON — remove cached file
		f.Close()
		os.Remove(path)
		return nil, fmt.Errorf("mtgjson: corrupt cache file %s: %w", filepath.Base(path), err)
	}
	return result, nil
}

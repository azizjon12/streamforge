package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	BaseDir string // ./hls
}

func NewLocalStorage(baseDir string) *LocalStorage {
	return &LocalStorage{BaseDir: baseDir}
}

func (s *LocalStorage) EnsurePrefix(ctx context.Context, prefix string) error {
	// prefix is relative like "<id>"
	dir := filepath.Join(s.BaseDir, prefix)
	return os.MkdirAll(dir, 0o755)
}

func (s *LocalStorage) OutputDir(prefix string) string {
	return filepath.Join(s.BaseDir, prefix)
}

func (s *LocalStorage) PlaylistPath(prefix string) string {
	return fmt.Sprintf("/hls/%s/playlist.m3u8", prefix)
}

func CloudFrontPlaylistURL(domain, prefix, id string) string {
	domain = strings.TrimSpace(domain)
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.Trim(prefix, "/")

	if prefix == "" {
		return "https://" + domain + "/" + id + "/playlist.m3u8"
	}

	return "https://" + domain + "/" + prefix + "/" + id + "/playlist.m3u8"
}

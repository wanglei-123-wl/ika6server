package play

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Deployment struct {
	GameID       int64  `json:"gameId"`
	RootPath     string `json:"rootPath"`
	EntryPath    string `json:"entryPath"`
	PublicURL    string `json:"publicUrl"`
	Status       string `json:"status"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

type Service struct {
	rootDir string
}

func NewService(rootDir string) *Service {
	return &Service{rootDir: rootDir}
}

func (s *Service) DeployZip(ctx context.Context, gameID int64, zipPath string) (Deployment, error) {
	if gameID <= 0 || strings.TrimSpace(zipPath) == "" {
		return Deployment{}, errors.New("game and zip file are required")
	}
	if strings.ToLower(filepath.Ext(zipPath)) != ".zip" {
		return Deployment{}, errors.New("online play build must be a .zip file")
	}
	gameRoot := filepath.Join(s.rootDir, strconv.FormatInt(gameID, 10))
	if err := os.RemoveAll(gameRoot); err != nil {
		return Deployment{}, err
	}
	if err := os.MkdirAll(gameRoot, 0755); err != nil {
		return Deployment{}, err
	}
	entry, err := extractZip(ctx, zipPath, gameRoot)
	if err != nil {
		_ = os.RemoveAll(gameRoot)
		return Deployment{}, err
	}
	return Deployment{
		GameID:    gameID,
		RootPath:  gameRoot,
		EntryPath: entry,
		PublicURL: "/play/games/" + strconv.FormatInt(gameID, 10) + "/" + path.Clean(strings.ReplaceAll(entry, `\`, "/")),
		Status:    "ready",
	}, nil
}

func (s *Service) Resolve(gameID int64, requestPath string) (string, error) {
	if gameID <= 0 {
		return "", errors.New("game id is required")
	}
	cleanPath, err := cleanArchivePath(requestPath)
	if err != nil {
		return "", err
	}
	root := filepath.Join(s.rootDir, strconv.FormatInt(gameID, 10))
	target := filepath.Join(root, filepath.FromSlash(cleanPath))
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", errors.New("invalid play asset path")
	}
	return target, nil
}

func extractZip(ctx context.Context, zipPath, destination string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	entryPath := ""
	for _, item := range reader.File {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		cleanName, err := cleanArchivePath(item.Name)
		if err != nil {
			return "", err
		}
		target := filepath.Join(destination, filepath.FromSlash(cleanName))
		rel, err := filepath.Rel(destination, target)
		if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
			return "", errors.New("archive contains unsafe path")
		}
		if item.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return "", err
		}
		if err := extractFile(item, target); err != nil {
			return "", err
		}
		if strings.EqualFold(path.Base(cleanName), "index.html") && entryPath == "" {
			entryPath = cleanName
		}
	}
	if entryPath == "" {
		return "", errors.New("online play build must contain index.html")
	}
	return entryPath, nil
}

func extractFile(item *zip.File, target string) error {
	src, err := item.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, item.Mode())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func cleanArchivePath(name string) (string, error) {
	name = strings.ReplaceAll(strings.TrimSpace(name), `\`, "/")
	name = path.Clean(name)
	if name == "." || strings.HasPrefix(name, "../") || strings.HasPrefix(name, "/") || strings.Contains(name, ":") {
		return "", fmt.Errorf("archive contains unsafe path %q", name)
	}
	return name, nil
}

func (s *Service) Remove(gameID int64) error {
	if gameID <= 0 {
		return nil
	}
	return os.RemoveAll(filepath.Join(s.rootDir, strconv.FormatInt(gameID, 10)))
}

func NowUTC() time.Time {
	return time.Now().UTC()
}

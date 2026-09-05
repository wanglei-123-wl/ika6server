package files

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wanglei-123-wl/ika6server/backend/internal/blocklist"
	"github.com/wanglei-123-wl/ika6server/backend/internal/sandbox"
	"github.com/wanglei-123-wl/ika6server/backend/internal/scanner"
)

type File struct {
	ID           int64          `json:"id"`
	PostID       int64          `json:"postId"`
	OriginalName string         `json:"originalName"`
	StoredName   string         `json:"storedName"`
	Size         int64          `json:"size"`
	SHA256       string         `json:"sha256"`
	Scan         scanner.Result `json:"scan"`
	Sandbox      sandbox.Report `json:"sandbox"`
	CreatedAt    time.Time      `json:"createdAt"`
}

type Store struct {
	mu        sync.RWMutex
	nextID    int64
	uploadDir string
	tempDir   string
	scanner   *scanner.Scanner
	sandbox   *sandbox.Analyzer
	blocklist *blocklist.Store
	byPost    map[int64]File
}

func NewStore(uploadDir, tempDir string, scanner *scanner.Scanner, sandbox *sandbox.Analyzer, blocklist *blocklist.Store) *Store {
	return &Store{
		nextID:    1,
		uploadDir: uploadDir,
		tempDir:   tempDir,
		scanner:   scanner,
		sandbox:   sandbox,
		blocklist: blocklist,
		byPost:    make(map[int64]File),
	}
}

func (s *Store) Save(ctx context.Context, postID int64, header *multipart.FileHeader) (File, error) {
	if postID <= 0 || header == nil {
		return File{}, errors.New("post and file are required")
	}

	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return File{}, err
	}
	if err := os.MkdirAll(s.tempDir, 0755); err != nil {
		return File{}, err
	}

	src, err := header.Open()
	if err != nil {
		return File{}, err
	}
	defer src.Close()

	original := filepath.Base(header.Filename)
	if original == "." || original == string(filepath.Separator) {
		return File{}, errors.New("invalid file name")
	}

	stored := strings.ReplaceAll(time.Now().UTC().Format("20060102150405.000000000"), ".", "") + "-" + original
	tempPath := filepath.Join(s.tempDir, stored)
	dstPath := filepath.Join(s.uploadDir, stored)
	dst, err := os.Create(tempPath)
	if err != nil {
		return File{}, err
	}

	size, err := io.Copy(dst, src)
	closeErr := dst.Close()
	if err != nil {
		_ = os.Remove(tempPath)
		return File{}, err
	}
	if closeErr != nil {
		_ = os.Remove(tempPath)
		return File{}, closeErr
	}

	scan, err := s.scanner.ScanFile(ctx, tempPath)
	if err != nil {
		_ = os.Remove(tempPath)
		return File{}, err
	}
	if !scan.Clean {
		_ = os.Remove(tempPath)
		return File{}, errors.New("file rejected by malware scan")
	}
	if entry, blocked := s.blocklist.Contains(scan.SHA256); blocked {
		_ = os.Remove(tempPath)
		return File{}, errors.New("file rejected by download blocklist: " + entry.Reason)
	}

	sandboxReport, err := s.sandbox.Analyze(ctx, tempPath)
	if err != nil {
		_ = os.Remove(tempPath)
		return File{}, err
	}

	if err := os.Rename(tempPath, dstPath); err != nil {
		_ = os.Remove(tempPath)
		return File{}, err
	}

	file := File{
		PostID:       postID,
		OriginalName: original,
		StoredName:   stored,
		Size:         size,
		SHA256:       scan.SHA256,
		Scan:         scan,
		Sandbox:      sandboxReport,
		CreatedAt:    time.Now().UTC(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	file.ID = s.nextID
	s.nextID++
	s.byPost[postID] = file
	return file, nil
}

func (s *Store) FindByPost(postID int64) (File, string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, ok := s.byPost[postID]
	if !ok {
		return File{}, "", false
	}

	return file, filepath.Join(s.uploadDir, file.StoredName), true
}

package files

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type File struct {
	ID           int64     `json:"id"`
	PostID       int64     `json:"postId"`
	OriginalName string    `json:"originalName"`
	StoredName   string    `json:"storedName"`
	Size         int64     `json:"size"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Store struct {
	mu        sync.RWMutex
	nextID    int64
	uploadDir string
	byPost    map[int64]File
}

func NewStore(uploadDir string) *Store {
	return &Store{
		nextID:    1,
		uploadDir: uploadDir,
		byPost:    make(map[int64]File),
	}
}

func (s *Store) Save(postID int64, header *multipart.FileHeader) (File, error) {
	if postID <= 0 || header == nil {
		return File{}, errors.New("post and file are required")
	}

	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
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
	dstPath := filepath.Join(s.uploadDir, stored)
	dst, err := os.Create(dstPath)
	if err != nil {
		return File{}, err
	}
	defer dst.Close()

	size, err := io.Copy(dst, src)
	if err != nil {
		return File{}, err
	}

	file := File{
		PostID:       postID,
		OriginalName: original,
		StoredName:   stored,
		Size:         size,
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

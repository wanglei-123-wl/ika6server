package scanner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	ClamScanBin string
	ClamAVDBDir string
	SevenZipBin string
	YaraBin     string
	YaraRules   string
}

type Result struct {
	Clean     bool     `json:"clean"`
	SHA256    string   `json:"sha256"`
	Engine    string   `json:"engine"`
	Findings  []string `json:"findings"`
	Warnings  []string `json:"warnings"`
	ScannedAt string   `json:"scannedAt"`
	ElapsedMS int64    `json:"elapsedMs"`
}

type Scanner struct {
	config Config
}

func New(config Config) *Scanner {
	return &Scanner{config: config}
}

func (s *Scanner) ScanFile(ctx context.Context, filePath string) (Result, error) {
	started := time.Now()
	result := Result{
		Clean:     true,
		Engine:    "clamav+yara",
		ScannedAt: started.UTC().Format(time.RFC3339),
	}

	hash, err := fileSHA256(filePath)
	if err != nil {
		return result, err
	}
	result.SHA256 = hash

	if err := s.runClamAV(ctx, filePath, &result); err != nil {
		return result, err
	}

	if err := s.runYARA(ctx, filePath, &result); err != nil {
		return result, err
	}
	if isArchive(filePath) {
		if err := s.scanArchiveContents(ctx, filePath, &result); err != nil {
			return result, err
		}
	}

	result.ElapsedMS = time.Since(started).Milliseconds()
	return result, nil
}

func (s *Scanner) scanArchiveContents(ctx context.Context, filePath string, result *Result) error {
	if s.config.SevenZipBin == "" {
		return errors.New("7-zip path is not configured")
	}
	if _, err := os.Stat(s.config.SevenZipBin); err != nil {
		return fmt.Errorf("7-zip is unavailable: %w", err)
	}

	extractDir, err := os.MkdirTemp("", "ika6-scan-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(extractDir)

	cmd := exec.CommandContext(ctx, s.config.SevenZipBin, "x", "-y", "-bd", "-o"+extractDir, filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("archive extraction failed: %s", strings.TrimSpace(string(output)))
	}

	if err := s.runClamAV(ctx, extractDir, result); err != nil {
		return err
	}

	return s.runYARA(ctx, extractDir, result)
}

func (s *Scanner) runClamAV(ctx context.Context, filePath string, result *Result) error {
	if s.config.ClamScanBin == "" {
		return errors.New("clamscan path is not configured")
	}
	if _, err := os.Stat(s.config.ClamScanBin); err != nil {
		return fmt.Errorf("clamscan is unavailable: %w", err)
	}

	args := []string{"--no-summary", "--recursive=yes"}
	if s.config.ClamAVDBDir != "" {
		args = append(args, "--database="+s.config.ClamAVDBDir)
	}
	args = append(args, filePath)

	cmd := exec.CommandContext(ctx, s.config.ClamScanBin, args...)
	output, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(output))

	if err == nil {
		return nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		result.Clean = false
		result.Findings = append(result.Findings, lines(text)...)
		return nil
	}

	return fmt.Errorf("clamav scan failed: %s", text)
}

func (s *Scanner) runYARA(ctx context.Context, filePath string, result *Result) error {
	if s.config.YaraBin == "" || s.config.YaraRules == "" {
		result.Warnings = append(result.Warnings, "yara skipped: path or rules not configured")
		return nil
	}
	if _, err := os.Stat(s.config.YaraBin); err != nil {
		result.Warnings = append(result.Warnings, "yara skipped: executable unavailable")
		return nil
	}
	if !hasYaraRules(s.config.YaraRules) {
		result.Warnings = append(result.Warnings, "yara skipped: no rules found")
		return nil
	}

	cmd := exec.CommandContext(ctx, s.config.YaraBin, "-r", s.config.YaraRules, filePath)
	output, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(output))

	if err == nil {
		return nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		result.Clean = false
		result.Findings = append(result.Findings, lines(text)...)
		return nil
	}

	return fmt.Errorf("yara scan failed: %s", text)
}

func fileSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func hasYaraRules(root string) bool {
	seen := false
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".yar" || ext == ".yara" {
			seen = true
		}
		return nil
	})
	return seen
}

func isArchive(filePath string) bool {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".zip", ".7z", ".rar", ".tar", ".gz", ".tgz", ".bz2", ".xz":
		return true
	default:
		return false
	}
}

func lines(text string) []string {
	if text == "" {
		return nil
	}

	raw := strings.Split(text, "\n")
	result := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

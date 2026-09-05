package sandbox

import (
	"archive/zip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Report struct {
	Mode        string    `json:"mode"`
	Risk        string    `json:"risk"`
	Findings    []Finding `json:"findings"`
	AnalyzedAt  time.Time `json:"analyzedAt"`
	Executable  bool      `json:"executable"`
	ArchiveRead bool      `json:"archiveRead"`
}

type Finding struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
	Path    string `json:"path"`
}

type Analyzer struct{}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) Analyze(ctx context.Context, filePath string) (Report, error) {
	report := Report{
		Mode:       "static-sandbox-gate",
		Risk:       "low",
		AnalyzedAt: time.Now().UTC(),
	}

	if strings.EqualFold(filepath.Ext(filePath), ".zip") {
		report.ArchiveRead = true
		if err := inspectZip(ctx, filePath, &report); err != nil {
			return report, err
		}
		return report, nil
	}

	if err := inspectFile(ctx, filePath, filepath.Base(filePath), &report); err != nil {
		return report, err
	}
	return report, nil
}

func inspectZip(ctx context.Context, filePath string, report *Report) error {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, item := range reader.File {
		if err := ctx.Err(); err != nil {
			return err
		}
		if item.FileInfo().IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(item.Name))
		if executableExt(ext) {
			report.Executable = true
			report.Findings = append(report.Findings, Finding{
				Rule:    "archive_contains_executable",
				Message: "archive contains an executable or script file",
				Path:    item.Name,
			})
		}

		file, err := item.Open()
		if err != nil {
			return err
		}
		err = inspectReader(ctx, file, item.Name, report)
		_ = file.Close()
		if err != nil {
			return err
		}
	}

	updateRisk(report)
	return nil
}

func inspectFile(ctx context.Context, filePath, name string, report *Report) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(name))
	if executableExt(ext) {
		report.Executable = true
		report.Findings = append(report.Findings, Finding{
			Rule:    "uploaded_executable",
			Message: "uploaded file is executable or script-like",
			Path:    name,
		})
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := inspectReader(ctx, file, name, report); err != nil {
		return err
	}
	updateRisk(report)
	return nil
}

func inspectReader(ctx context.Context, reader io.Reader, name string, report *Report) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	limited, err := io.ReadAll(io.LimitReader(reader, 2<<20))
	if err != nil {
		return err
	}
	text := strings.ToLower(string(limited))

	checks := []struct {
		rule    string
		needle  string
		message string
	}{
		{"network_download", "invoke-webrequest", "file contains PowerShell download command"},
		{"network_download", "downloadstring", "file contains download-and-load pattern"},
		{"shell_execution", "invoke-expression", "file contains dynamic execution command"},
		{"shell_execution", "system.net.webclient", "file contains web client scripting"},
		{"persistence", "currentversion\\run", "file references Windows startup persistence key"},
		{"credential_access", "mimikatz", "file references a known credential theft tool name"},
	}

	for _, check := range checks {
		if strings.Contains(text, check.needle) {
			report.Findings = append(report.Findings, Finding{
				Rule:    check.rule,
				Message: check.message,
				Path:    name,
			})
		}
	}
	return nil
}

func updateRisk(report *Report) {
	if len(report.Findings) == 0 {
		report.Risk = "low"
		return
	}
	report.Risk = "medium"
	for _, finding := range report.Findings {
		if finding.Rule == "credential_access" || finding.Rule == "persistence" {
			report.Risk = "high"
			return
		}
	}
}

func executableExt(ext string) bool {
	switch ext {
	case ".exe", ".dll", ".bat", ".cmd", ".ps1", ".vbs", ".js", ".msi", ".scr", ".com":
		return true
	default:
		return false
	}
}

var ErrUnsafeToRun = errors.New("uploaded code is never executed in the app process")

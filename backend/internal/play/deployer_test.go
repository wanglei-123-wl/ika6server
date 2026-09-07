package play

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeployZipPublishesIndex(t *testing.T) {
	root := t.TempDir()
	archive := filepath.Join(t.TempDir(), "build.zip")
	writeZip(t, archive, map[string]string{"dist/index.html": "<h1>play</h1>", "dist/app.js": "console.log('ok')"})

	deployment, err := NewService(root).DeployZip(context.Background(), 7, archive)
	if err != nil {
		t.Fatal(err)
	}
	if deployment.Status != "ready" || deployment.EntryPath != "dist/index.html" || deployment.PublicURL != "/play/games/7/dist/index.html" {
		t.Fatalf("unexpected deployment: %#v", deployment)
	}
	if _, err := os.Stat(filepath.Join(root, "7", "dist", "index.html")); err != nil {
		t.Fatal(err)
	}
}

func TestDeployZipRejectsMissingIndex(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "build.zip")
	writeZip(t, archive, map[string]string{"asset.txt": "missing entry"})

	_, err := NewService(t.TempDir()).DeployZip(context.Background(), 1, archive)
	if err == nil || !strings.Contains(err.Error(), "index.html") {
		t.Fatalf("expected missing index error, got %v", err)
	}
}

func TestDeployZipRejectsUnsafePath(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "build.zip")
	writeZip(t, archive, map[string]string{"../index.html": "<h1>bad</h1>"})

	_, err := NewService(t.TempDir()).DeployZip(context.Background(), 1, archive)
	if err == nil || !strings.Contains(err.Error(), "unsafe path") {
		t.Fatalf("expected unsafe path error, got %v", err)
	}
}

func TestResolveRejectsUnsafePath(t *testing.T) {
	_, err := NewService(t.TempDir()).Resolve(1, "../secret.txt")
	if err == nil {
		t.Fatal("expected unsafe path to fail")
	}
}

func writeZip(t *testing.T, archive string, files map[string]string) {
	t.Helper()
	target, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(target)
	for name, content := range files {
		item, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := item.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := target.Close(); err != nil {
		t.Fatal(err)
	}
}

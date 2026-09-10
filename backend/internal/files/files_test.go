package files

import (
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateUpload(t *testing.T) {
	tests := []struct {
		name    string
		kind    string
		header  *multipart.FileHeader
		wantErr bool
	}{
		{name: "cover accepts image", kind: "cover", header: &multipart.FileHeader{Filename: "cover.png", Size: 1024}},
		{name: "cover rejects archive", kind: "cover", header: &multipart.FileHeader{Filename: "cover.zip", Size: 1024}, wantErr: true},
		{name: "build accepts zip", kind: "build", header: &multipart.FileHeader{Filename: "build.zip", Size: 1024}},
		{name: "build rejects non-zip", kind: "build", header: &multipart.FileHeader{Filename: "build.tar", Size: 1024}, wantErr: true},
		{name: "package accepts zip", kind: "source", header: &multipart.FileHeader{Filename: "source.zip", Size: 1024}},
		{name: "package rejects script", kind: "source", header: &multipart.FileHeader{Filename: "source.ps1", Size: 1024}, wantErr: true},
		{name: "empty rejects", kind: "source", header: &multipart.FileHeader{Filename: "source.zip", Size: 0}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateUpload(test.kind, test.header); (err != nil) != test.wantErr {
				t.Fatalf("validateUpload error = %v, wantErr = %v", err, test.wantErr)
			}
		})
	}
}

func TestRemoveOwnerKindsKeepsUnrelatedOwnerFiles(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir, dir, nil, nil, nil)
	for _, name := range []string{"cover.bin", "build.bin", "source.bin", "avatar.bin", "forum.bin"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0644); err != nil {
			t.Fatal(err)
		}
	}
	store.byKey["7:cover"] = File{StoredName: "cover.bin"}
	store.byKey["7:build"] = File{StoredName: "build.bin"}
	store.byKey["7:source"] = File{StoredName: "source.bin"}
	store.byKey["7:avatar"] = File{StoredName: "avatar.bin"}
	store.byKey["7:forum"] = File{StoredName: "forum.bin"}

	store.RemoveOwnerKinds(7, "cover", "build", "source")

	for _, name := range []string{"cover.bin", "build.bin", "source.bin"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("%s was not removed, err = %v", name, err)
		}
	}
	for _, name := range []string{"avatar.bin", "forum.bin"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("%s should be kept, err = %v", name, err)
		}
	}
	if _, ok := store.byKey["7:avatar"]; !ok {
		t.Fatal("avatar key should be kept")
	}
	if _, ok := store.byKey["7:forum"]; !ok {
		t.Fatal("forum key should be kept")
	}
}

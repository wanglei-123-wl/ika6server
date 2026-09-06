package files

import (
	"mime/multipart"
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

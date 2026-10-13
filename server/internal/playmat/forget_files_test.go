package playmat

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestForgetFilesRemovesImagesAndIgnoresStrangers(t *testing.T) {
	dir := t.TempDir()
	files := NewFileStore(dir)
	svc := NewService(files, nil, nil)
	a, b := uuid.NewString(), uuid.NewString()
	for _, id := range []string{a, b} {
		if err := os.WriteFile(filepath.Join(dir, id+".jpg"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := svc.ForgetFiles([]string{a, "../escape", "not-an-id"}); err != nil {
		t.Fatal(err)
	}
	if files.Exists(a) {
		t.Error("the named image survived")
	}
	if !files.Exists(b) {
		t.Error("an image that was not named was removed")
	}
	var nilSvc *Service
	if err := nilSvc.ForgetFiles([]string{a}); err != nil {
		t.Errorf("nil service: %v", err)
	}
}

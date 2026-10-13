package discord

import (
	"os"
	"path/filepath"
	"testing"
)

func TestForgetRemovesOnlyThatIDsAvatars(t *testing.T) {
	dir := t.TempDir()
	for _, p := range []string{"111/aaa.png", "111/bbb.png", "222/ccc.png"} {
		full := filepath.Join(dir, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, pngBytes, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	c := NewAvatarCache(dir, nil)
	if err := c.Forget("111"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "111")); !os.IsNotExist(err) {
		t.Errorf("the id's directory survived: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "222", "ccc.png")); err != nil {
		t.Errorf("another id's avatar was removed: %v", err)
	}
	// A key that could climb out of the cache is ignored, and so is a
	// cache with no directory.
	for _, bad := range []string{"", "..", "../222", "a/b"} {
		if err := c.Forget(bad); err != nil {
			t.Errorf("Forget(%q) = %v", bad, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "222", "ccc.png")); err != nil {
		t.Errorf("a malformed key removed something: %v", err)
	}
	if err := NewAvatarCache("", nil).Forget("111"); err != nil {
		t.Errorf("no-directory cache: %v", err)
	}
}

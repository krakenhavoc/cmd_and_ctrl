package ws

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestScrubPlayersBlanksOnlyTheNamedPlayer(t *testing.T) {
	gone, kept := uuid.NewString(), uuid.NewString()
	line := `{"seq":9007199254740993,"game":{"seats":[` +
		`{"id":"` + gone + `","name":"Alice","discord_id":"111","discord_avatar_hash":"abc","display_name":"Alice A","playmat_url":"/playmats/x","life":40},` +
		`{"id":"` + kept + `","name":"Bob","discord_id":"222","life":38}],` +
		`"shields":[{"id":"s1","target_id":"` + gone + `","target_name":"Alice","amount":2}]},` +
		`"annotation":{"seat":"` + gone + `","seat_name":"Alice","tag":"x"},` +
		`"snapshot":{"players":[{"id":"` + gone + `","name":"Alice","discordId":"111","discordAvatarHash":"abc","displayName":"Alice A"}]}}`

	out, changed, err := scrubLine([]byte(line), map[string]bool{gone: true}, "Deleted player")
	if err != nil || !changed {
		t.Fatalf("scrubLine = changed %v, err %v", changed, err)
	}
	got := string(out)
	for _, leak := range []string{"Alice", "111", `"abc"`, "/playmats/x", "discordId", "displayName"} {
		if strings.Contains(got, leak) {
			t.Errorf("output still contains %q:\n%s", leak, got)
		}
	}
	for _, want := range []string{"Deleted player", `"Bob"`, `"222"`, `"life":38`, `"amount":2`, "9007199254740993"} {
		if !strings.Contains(got, want) {
			t.Errorf("output lost %q:\n%s", want, got)
		}
	}
}

func TestScrubFileLeavesAnUntouchedFileAloneAndKeepsLineBreaks(t *testing.T) {
	dir := t.TempDir()
	id := uuid.NewString()
	other := uuid.NewString()

	untouched := filepath.Join(dir, "a.jsonl")
	const body = "{\"id\":\"" + "x" + "\",\"name\":\"Zed\"}\n"
	if err := os.WriteFile(untouched, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := scrubFile(untouched, map[string]bool{id: true}, "Deleted player"); err != nil {
		t.Fatal(err)
	}
	if raw, _ := os.ReadFile(untouched); string(raw) != body {
		t.Errorf("a file with nothing to scrub was rewritten: %q", raw)
	}

	touched := filepath.Join(dir, "b.jsonl")
	in := `{"id":"` + id + `","name":"Alice"}` + "\n" + `{"id":"` + other + `","name":"Bob"}` + "\n"
	if err := os.WriteFile(touched, []byte(in), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := scrubFile(touched, map[string]bool{id: true}, "Deleted player"); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(touched)
	lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	if len(lines) != 2 || !strings.Contains(lines[0], "Deleted player") || !strings.Contains(lines[1], "Bob") {
		t.Errorf("rewritten file = %q", raw)
	}
	if !strings.HasSuffix(string(raw), "\n") {
		t.Error("the trailing newline was lost: the next append would join two lines")
	}
	if info, _ := os.Stat(touched); info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %v, want 0600 kept", info.Mode().Perm())
	}

	bad := filepath.Join(dir, "c.jsonl")
	if err := os.WriteFile(bad, []byte("not json\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := scrubFile(bad, map[string]bool{id: true}, "x"); err == nil {
		t.Error("a line that is not JSON must be an error")
	}
	if raw, _ := os.ReadFile(bad); string(raw) != "not json\n" {
		t.Errorf("a failed scrub changed the file: %q", raw)
	}
	if err := scrubFile(filepath.Join(dir, "missing.jsonl"), map[string]bool{id: true}, "x"); err != nil {
		t.Errorf("a missing file is fine: %v", err)
	}
}

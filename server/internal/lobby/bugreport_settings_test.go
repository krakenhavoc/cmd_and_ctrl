package lobby

// bugreport_settings_test.go — an idea attaches a copy of the
// reporter's settings (#4103). The issue is public, so the copy holds
// preferences only (ADR 0149): no names, ids, tokens or game text.

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// reportSettings is a client settings object with something in every
// place the scrub has to handle.
func reportSettings() map[string]any {
	return map[string]any{
		"__version": 31,
		"display":   map[string]any{"cardSize": "large", "expandStyle": "overlay", "accent": "#c0ffee"},
		"audio":     map[string]any{"muted": false, "masterVolume": 80},
		"gameplay": map[string]any{
			"passMode": "careful",
			"autoAnswers": []any{map[string]any{
				"key":    "6b8cf2a0-b045-4d91-9d91-c602d40c6237|pay",
				"card":   "Rhystic Study",
				"prompt": "Pay {1} or Mallory Seatname draws a card",
				"answer": "always",
			}},
		},
		"shortcuts": map[string]any{"enabled": true, "bindings": map[string]any{"pass": "Space`"}},
		"help":      map[string]any{"tipsOff": true, "seen": map[string]any{"hint.first-game": 1}},
		// Not a Settings group: dropped whatever it holds.
		"session": map[string]any{"token": "abcdef0123456789"},
	}
}

// TestBugReportKindGatesTheSettingsCopy: only an idea takes the
// settings; a bug or a question that sends them has them dropped.
func TestBugReportKindGatesTheSettingsCopy(t *testing.T) {
	for _, tc := range []struct {
		kind string
		want bool
	}{
		{"idea", true},
		{"bug", false},
		{"question", false},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			rep := &recordingReporter{}
			srv, tok := newBugReportStack(t, rep)
			resp := postBugReport(t, srv, tok, map[string]any{
				"title": "wider stack panel", "kind": tc.kind, "settings": reportSettings(),
			})
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusCreated {
				t.Fatalf("status = %d, want 201", resp.StatusCode)
			}
			body := rep.bodies[0]
			if got := strings.Contains(body, "<summary>Settings"); got != tc.want {
				t.Errorf("settings block in body = %v, want %v\n----\n%s", got, tc.want, body)
			}
			if got := strings.Contains(body, `"cardSize": "large"`); got != tc.want {
				t.Errorf("settings values in body = %v, want %v\n----\n%s", got, tc.want, body)
			}
		})
	}
}

// TestBugSettingsCopyHoldsPreferencesOnly is the scrub: the groups
// stay, and what could name or follow a person goes.
func TestBugSettingsCopyHoldsPreferencesOnly(t *testing.T) {
	raw, err := json.Marshal(reportSettings())
	if err != nil {
		t.Fatal(err)
	}
	block, note := renderBugSettings(raw)
	if note != "" || block == "" {
		t.Fatalf("renderBugSettings = %q, note %q; want a block", block, note)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(block), &got); err != nil {
		t.Fatalf("the block is not JSON: %v\n%s", err, block)
	}
	for _, group := range []string{"__version", "display", "audio", "gameplay", "shortcuts", "help"} {
		if _, ok := got[group]; !ok {
			t.Errorf("group %q was dropped", group)
		}
	}
	for _, gone := range []string{"session", "abcdef0123456789", "Mallory", "Rhystic Study", "hint.first-game", "`"} {
		if strings.Contains(block, gone) {
			t.Errorf("the copy still holds %q:\n%s", gone, block)
		}
	}
	rules := got["gameplay"].(map[string]any)["autoAnswers"].([]any)
	if len(rules) != 1 {
		t.Fatalf("autoAnswers = %v, want the one rule", rules)
	}
	want := map[string]any{"key": "6b8cf2a0-b045-4d91-9d91-c602d40c6237|pay", "answer": "always"}
	if r := rules[0].(map[string]any); len(r) != 2 || r["key"] != want["key"] || r["answer"] != want["answer"] {
		t.Errorf("rule = %v, want only %v", r, want)
	}
	if h := got["help"].(map[string]any); len(h) != 1 || h["tipsOff"] != true {
		t.Errorf("help = %v, want only tipsOff", h)
	}
	if v := got["__version"]; v != float64(31) {
		t.Errorf("__version = %v, want 31", v)
	}
}

// A credential pasted anywhere in the object is redacted, the way the
// rest of the issue is (#721).
func TestBugSettingsCopyIsRedacted(t *testing.T) {
	raw := []byte(`{"display":{"accent":"https://cmd.labxp.io/#/join/x?token=s3cr3t-value-here"}}`)
	block, note := renderBugSettings(raw)
	if note != "" {
		t.Fatalf("note = %q", note)
	}
	if strings.Contains(block, "s3cr3t-value-here") {
		t.Errorf("a token survived the scrub:\n%s", block)
	}
}

// A copy that is not a settings object, or will not fit, is left out,
// and the issue says why instead of carrying half a copy.
func TestBugSettingsCopyLeftOutWhenItCannotBeAttached(t *testing.T) {
	big := map[string]any{}
	for i := range 400 {
		big[strings.Repeat("k", 40)+string(rune('a'+i%26))+strings.Repeat("x", i)] = strings.Repeat("v", 60)
	}
	bigRaw, _ := json.Marshal(map[string]any{"display": big})
	for _, tc := range []struct {
		name string
		raw  string
		note string
	}{
		{"none", "", ""},
		{"an array", `[1,2]`, "not a settings object"},
		{"a string", `"hi"`, "not a settings object"},
		{"too deep", `{"a":{"b":{"c":{"d":{"e":1}}}}}`, "not a settings object"},
		{"over the account's 32 KiB", string(bigRaw), "not a settings object"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			block, note := renderBugSettings(json.RawMessage(tc.raw))
			if block != "" || note != tc.note {
				t.Errorf("renderBugSettings = %q, note %q; want no block, note %q", block, note, tc.note)
			}
		})
	}

	// Within the account's cap but over the rendered cap: compact is
	// tried, and if that is still too long the copy is left out.
	fits := map[string]any{}
	for i := range 220 {
		fits[strings.Repeat("k", 30)+string(rune('a'+i%26))+strings.Repeat("y", i/26)] = strings.Repeat("v", 90)
	}
	raw, _ := json.Marshal(map[string]any{"display": fits})
	if len(raw) > 32<<10 || len(raw) <= bugSettingsTotalMax {
		t.Fatalf("fixture is %d bytes; want between %d and 32 KiB", len(raw), bugSettingsTotalMax)
	}
	if block, note := renderBugSettings(raw); block != "" || note != "over 24 KiB" {
		t.Errorf("renderBugSettings = %d bytes, note %q; want no block, note %q", len(block), note, "over 24 KiB")
	}

	rep := &recordingReporter{}
	srv, tok := newBugReportStack(t, rep)
	resp := postBugReport(t, srv, tok, map[string]any{"title": "x", "kind": "idea", "settings": []int{1}})
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201: a bad settings copy must not cost the report", resp.StatusCode)
	}
	if !strings.Contains(rep.bodies[0], "- Settings: not attached (not a settings object)") {
		t.Errorf("the body does not say the copy was left out:\n%s", rep.bodies[0])
	}
}

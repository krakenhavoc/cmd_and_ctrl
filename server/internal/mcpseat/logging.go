package mcpseat

import (
	"io"
	"log/slog"
	"strings"
	"sync"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/redact"
)

// secrets is the set of credential strings this process holds. Every
// log line and every tool result passes through scrub, which replaces
// each of them, after redact.Secrets has done its pattern work. The
// pattern rules alone would miss a bare session token printed on its
// own, so the exact strings are replaced too (§8: "never in a tool
// result", "every log line passes through redact.Secrets").
type secrets struct {
	mu   sync.RWMutex
	vals map[string]struct{}
}

func newSecrets() *secrets { return &secrets{vals: map[string]struct{}{}} }

// add registers a credential. Short strings are ignored: an empty or
// tiny value would replace ordinary text.
func (s *secrets) add(v string) {
	if len(v) < 8 {
		return
	}
	s.mu.Lock()
	s.vals[v] = struct{}{}
	s.mu.Unlock()
}

// scrub removes every known credential and every pattern redact knows.
func (s *secrets) scrub(text string) string {
	if text == "" {
		return text
	}
	s.mu.RLock()
	for v := range s.vals {
		if strings.Contains(text, v) {
			text = strings.ReplaceAll(text, v, redact.Redacted)
		}
	}
	s.mu.RUnlock()
	return redact.Secrets(text)
}

// scrubWriter applies scrub to everything written through it. slog's
// handlers write one record per Write call, so a line is never split
// across two scrubs.
type scrubWriter struct {
	mu  sync.Mutex
	w   io.Writer
	sec *secrets
}

func (sw *scrubWriter) Write(p []byte) (int, error) {
	out := sw.sec.scrub(string(p))
	sw.mu.Lock()
	defer sw.mu.Unlock()
	if _, err := io.WriteString(sw.w, out); err != nil {
		return 0, err
	}
	return len(p), nil
}

// NewLogger returns a text logger on w whose every line is scrubbed of
// the secrets in sec. stdout carries the MCP protocol, so w is stderr or
// the --log-file, never stdout.
func newLogger(w io.Writer, sec *secrets, level slog.Level) *slog.Logger {
	return slog.New(slog.NewTextHandler(&scrubWriter{w: w, sec: sec}, &slog.HandlerOptions{Level: level}))
}

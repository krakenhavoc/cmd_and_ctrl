package docsguard_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// serverRoot is server/ relative to this package.
const serverRoot = "../.."

// gatewayAllowlist names the handler files that may write a 502 or 504,
// each with the reason. It is empty on purpose (#1644).
//
// Cloudflare sits in front of cmd-dev and prod and REPLACES the body of
// any 502 or 504 the origin returns with its own "error code: 502"
// page, so the JSON `error` sentence and `hint` a player needs never
// arrive. An upstream failure answers 424 Failed Dependency instead,
// which Cloudflare passes through (#1643, #1644). A file goes here only
// for a status that is genuinely a gateway's, with the reason.
var gatewayAllowlist = map[string]string{}

// TestHandlersDoNotAnswer502Or504 fails when production Go code under
// server/ names http.StatusBadGateway or http.StatusGatewayTimeout, or
// hands WriteHeader a literal 502 / 504.
func TestHandlersDoNotAnswer502Or504(t *testing.T) {
	banned := []string{"StatusBadGateway", "StatusGatewayTimeout", "WriteHeader(502)", "WriteHeader(504)"}
	err := filepath.WalkDir(serverRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, _ := filepath.Rel(serverRoot, path)
		rel = filepath.ToSlash(rel)
		if _, ok := gatewayAllowlist[rel]; ok {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, w := range banned {
			if strings.Contains(string(b), w) {
				t.Errorf("%s uses %s: Cloudflare replaces a 502/504 body, so the player never sees the error. Answer http.StatusFailedDependency (424), or add the file to gatewayAllowlist with a reason", rel, w)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for f, reason := range gatewayAllowlist {
		if reason == "" {
			t.Errorf("gatewayAllowlist[%q] has no reason", f)
		}
	}
}

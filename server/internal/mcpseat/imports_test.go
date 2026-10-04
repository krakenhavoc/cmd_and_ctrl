package mcpseat

import (
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// imports_test.go is ADR 0122 §2's type gate, and §1's rule that the SDK
// lives in one file.
//
// The seat runs on the owner's workstation with a socket to a server
// elsewhere, so there is no *game.Game in its process to read whatever it
// imports. The gate keeps the code saying the same thing: the binary sees
// only what the wire gives its seat. It bans the DIRECT import only, as
// the aiseat gate does: protocol and legal import internal/game
// themselves, so a transitive ban is not expressible, and a linked
// package is not a handle on the table.

const modulePrefix = "github.com/krakenhavoc/cmd_and_ctrl/server/"

var bannedServerState = []string{
	modulePrefix + "internal/game",
	modulePrefix + "internal/ws",
	modulePrefix + "internal/lobby",
	modulePrefix + "internal/actions",
	modulePrefix + "internal/db",
	modulePrefix + "internal/auth",
	modulePrefix + "internal/aiseat/tiers",
}

const sdkPrefix = "github.com/modelcontextprotocol/go-sdk"

// gatedDirs are the binary's packages, relative to this one.
var gatedDirs = []string{".", "../../cmd/mcpseat"}

// TestMCPSeatImportsNoServerState checks Imports and in-package
// TestImports. XTestImports is exempt: the end-to-end test (e2e_test.go,
// package mcpseat_test) has to stand a real server up, and a test that
// stands the server up is the one place the server's packages belong.
func TestMCPSeatImportsNoServerState(t *testing.T) {
	for _, dir := range gatedDirs {
		pkg, err := build.ImportDir(dir, 0)
		if err != nil {
			t.Fatalf("import %s: %v", dir, err)
		}
		for label, imports := range map[string][]string{"": pkg.Imports, "test ": pkg.TestImports} {
			for _, imp := range imports {
				for _, banned := range bannedServerState {
					if imp == banned {
						t.Errorf("%s %simports %s — the MCP seat reads only the wire (ADR 0122 §2)", dir, label, imp)
					}
				}
			}
		}
	}
}

// TestServerStateBanIsLoadBearing is the control: if the module path
// changed, the gate above would pass vacuously. The server's own main
// imports most of the banned packages, so the list must match there.
func TestServerStateBanIsLoadBearing(t *testing.T) {
	pkg, err := build.ImportDir("../../cmd/server", 0)
	if err != nil {
		t.Fatalf("import cmd/server: %v", err)
	}
	matched := map[string]bool{}
	for _, imp := range pkg.Imports {
		for _, banned := range bannedServerState {
			if imp == banned {
				matched[banned] = true
			}
		}
	}
	for _, want := range []string{"internal/game", "internal/ws", "internal/lobby", "internal/auth", "internal/aiseat/tiers"} {
		if !matched[modulePrefix+want] {
			t.Errorf("cmd/server no longer imports %s: the ban list's paths have drifted from the module's", want)
		}
	}
}

// TestOnlyTransportImportsTheSDK holds §1: the SDK is imported by
// transport.go and nothing else in the binary (its test, transport_test.go,
// and the external e2e test drive it as a client, and are exempt).
func TestOnlyTransportImportsTheSDK(t *testing.T) {
	allowed := map[string]bool{"transport.go": true, "transport_test.go": true, "e2e_test.go": true}
	found := 0
	for _, dir := range gatedDirs {
		files, err := filepath.Glob(filepath.Join(dir, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			for _, imp := range fileImports(t, file) {
				if !strings.HasPrefix(imp, sdkPrefix) {
					continue
				}
				found++
				if !(dir == "." && allowed[filepath.Base(file)]) {
					t.Errorf("%s imports %s; only transport.go may (ADR 0122 §1)", file, imp)
				}
			}
		}
	}
	if found == 0 {
		t.Fatal("no file imports the SDK at all: the gate's prefix no longer matches transport.go")
	}
}

func fileImports(t *testing.T, path string) []string {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	out := make([]string, 0, len(f.Imports))
	for _, is := range f.Imports {
		p, err := strconv.Unquote(is.Path.Value)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, p)
	}
	return out
}

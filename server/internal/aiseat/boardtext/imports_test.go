package boardtext_test

import (
	"go/build"
	"strings"
	"testing"
)

// The renderer is shared by the bot and the MCP seat (ADR 0122 §5), so
// it may name nothing that holds authoritative state or reaches the
// network. heuristic/imports_test.go already bans internal/game from
// every aiseat subpackage; this adds the packages the MCP binary's own
// import gate (ADR 0122 §2) keeps out as well.
func TestBoardtextImportsStayPure(t *testing.T) {
	banned := []string{
		"/server/internal/game",
		"/server/internal/ws",
		"/server/internal/lobby",
		"/server/internal/actions",
	}
	pkg, err := build.ImportDir(".", 0)
	if err != nil {
		t.Fatal(err)
	}
	all := map[string][]string{"": pkg.Imports, "test ": pkg.TestImports, "xtest ": pkg.XTestImports}
	for label, imports := range all {
		for _, imp := range imports {
			for _, ban := range banned {
				if strings.HasSuffix(imp, ban) {
					t.Errorf("boardtext %simports %s", label, imp)
				}
			}
		}
	}
	if len(pkg.Imports) == 0 {
		t.Fatal("no imports found: the test is not looking at the package")
	}
}

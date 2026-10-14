package docsguard_test

import (
	"os"
	"testing"
)

// agentsFile is AGENTS.md relative to this package.
const agentsFile = "../../../AGENTS.md"

// agentsMaxBytes caps AGENTS.md at 15 KB. Every agent reads the file on
// every session, so the rules stay in it and the reference (commands,
// environment variables, endpoints, test knobs) goes to docs/ with a link.
const agentsMaxBytes = 15_000

// TestAgentsMDStaysSlim is #3351's guard: AGENTS.md is rules and links,
// not a reference manual. If it grows past the cap, move the new material
// to the docs/ file it belongs to and link it from AGENTS.md §3 or §5.
func TestAgentsMDStaysSlim(t *testing.T) {
	info, err := os.Stat(agentsFile)
	if err != nil {
		t.Fatalf("stat %s: %v", agentsFile, err)
	}
	if info.Size() > agentsMaxBytes {
		t.Errorf("AGENTS.md is %d bytes, over the %d-byte cap: move reference material to docs/ and link it from AGENTS.md", info.Size(), agentsMaxBytes)
	}
}

package effects

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bootProbeEnv marks the child process that runs only the boot probe.
const bootProbeEnv = "CMDCTRL_EFFECT_KEYS_BOOT_PROBE"

// TestEveryLedgerKeyResolvesAtBoot is #4274's check. A restore point can
// name any effect key in game/testdata/effect_keys.txt, and a freshly
// booted binary has run no card yet. A registry that fills its keys on
// first use (a sync.Once, a register call inside an Apply or a helper, a
// lazily filled map) passes TestEveryPersistedEffectKeyResolves whenever
// an earlier test happened to run the card, and then fails a real
// restore. So this test re-runs the test binary with ONLY the probe
// below selected: no card has played and no other test has touched a
// registry, so a key that resolves there resolves at boot.
func TestEveryLedgerKeyResolvesAtBoot(t *testing.T) {
	if os.Getenv(bootProbeEnv) != "" {
		t.Skip("child process: the probe runs, not this wrapper")
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run", "^TestLedgerKeysResolveBeforeAnyCardRuns$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(), bootProbeEnv+"=1")
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "--- PASS: TestLedgerKeysResolveBeforeAnyCardRuns") {
		t.Fatalf("a ledger key does not resolve in a freshly booted binary (a registry fills lazily; "+
			"register it at init, as drawThenDiscardRef does): %v\n%s", err, out)
	}
}

// TestLedgerKeysResolveBeforeAnyCardRuns is the child half. It is skipped
// in a normal run, so it can never run after another test.
func TestLedgerKeysResolveBeforeAnyCardRuns(t *testing.T) {
	if os.Getenv(bootProbeEnv) == "" {
		t.Skip("run by TestEveryLedgerKeyResolvesAtBoot in a child process")
	}
	path := filepath.Join("..", "..", "game", "testdata", "effect_keys.txt")
	missing, stale, err := game.CheckEffectKeyLedger(path, false, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range missing {
		t.Errorf("%q is registered at boot but not in the ledger", line)
	}
	for _, line := range stale {
		t.Errorf("ledger line %q does not resolve before any card has run", line)
	}
}

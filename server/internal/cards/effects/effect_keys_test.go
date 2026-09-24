package effects

import (
	"flag"
	"path/filepath"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

var updateEffectKeys = flag.Bool("update-effect-keys", false,
	"append newly registered effect keys to game/testdata/effect_keys.txt (never removes a line)")

// TestEveryPersistedEffectKeyResolves is the full ledger check (ADR
// 0041 phase 3, tier 2, #1497). This package has every production body
// and condition registered — the engine's and the catalog's — so it can
// say both halves: every registered key is in the ledger, and every
// ledger line still resolves. A key a restore point names must resolve
// in every later binary; keys are never deleted, only aliased.
func TestEveryPersistedEffectKeyResolves(t *testing.T) {
	path := filepath.Join("..", "..", "game", "testdata", "effect_keys.txt")
	missing, stale, err := game.CheckEffectKeyLedger(path, *updateEffectKeys, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range missing {
		t.Errorf(`%q is registered but not in game/testdata/effect_keys.txt.

Every effect key can end up in a restore point, so it is an on-disk
identity and the ledger has to know it. Append it:
  go test ./internal/cards/effects -run TestEveryPersistedEffectKeyResolves -args -update-effect-keys`, line)
	}
	for _, line := range stale {
		t.Errorf(`game/testdata/effect_keys.txt lists %q, which no longer resolves.

A restore point written by an earlier binary can name this key, and this
binary would refuse it. Keys are never deleted: register the new name and
keep this one with game.EffectAlias(old, new).`, line)
	}
}

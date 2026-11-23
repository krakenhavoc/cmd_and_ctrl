package effectstest

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// bloodthirsterTargets is who the enumerator offers `attacker` at, as
// seat ids.
func BloodthirsterTargets(t *testing.T, g *game.Game, seat, attacker uuid.UUID) map[string]bool {
	t.Helper()
	out := map[string]bool{}
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Kind != legal.KindAttack || m.Source != attacker {
			continue
		}
		var p struct {
			Target string `json:"target"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("attack params %s: %v", m.Params, err)
		}
		out[p.Target] = true
	}
	return out
}

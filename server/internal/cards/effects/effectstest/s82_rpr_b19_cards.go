package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rprB19Cast puts a spell in `who`'s hand, floats `mana` and casts it under
// the strict mana gate.
func RprB19Cast(t *testing.T, g *game.Game, who *game.Player, name, typeLine, cost, oracle, mana string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	id := HandCardFull(who, name, typeLine, cost, oracle, nil)
	if mana != "" {
		if err := g.AddManaForEffect(who.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	}
	params.Strict = true
	return id, g.CastSpell(who.ID, id, params)
}

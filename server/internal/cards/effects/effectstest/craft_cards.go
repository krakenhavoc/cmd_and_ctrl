package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// craft_cards_test.go — one test per card #2124 ships on the craft
// seam (ADR 0137). The seam itself is craft_test.go.

// craftInto activates `source`'s craft ability (its ability `index`)
// with `materials`, paying `mana`, and settles the stack. Returns the
// crafted permanent, found by its back face's name.
func CraftInto(t *testing.T, g *game.Game, p *game.Player, source uuid.UUID, index int, mana, back string, materials ...uuid.UUID) game.Card {
	t.Helper()
	FloatMana(t, g, p, mana)
	if err := ActivateCraft(g, p, source, index, materials...); err != nil {
		t.Fatalf("craft into %s: %v", back, err)
	}
	PassPriorityAroundTable(t, g)
	return CraftedPermanent(t, g, back)
}

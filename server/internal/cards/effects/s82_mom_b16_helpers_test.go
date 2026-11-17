package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82mom16Card builds a transform card as the importer does: two faces,
// the front materialised, the back's oracle registered under "#1".
func s82mom16Card(owner uuid.UUID, oracle string, front, back game.Face) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces:      []game.Face{front, back},
	}
	c.SetFace(0)
	return c
}

// s82mom16Transform pays for and activates ability 0 of the permanent
// (the Phyrexian transform) with the given mana, from a main phase,
// and settles the stack.
func s82mom16Transform(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, mana string, life int) {
	t.Helper()
	s82mom16TransformAbility(t, g, me, id, 0, mana, life)
}

// s82mom16TransformAbility is s82mom16Transform for the ability at
// index, `life` being how many Phyrexian symbols are paid with 2 life.
func s82mom16TransformAbility(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, index int, mana string, life int) {
	t.Helper()
	if mana != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, id, index, game.ActivateAbilityParams{PhyrexianLife: life, Strict: true}); err != nil {
		t.Fatalf("transform: %v", err)
	}
	passPriorityAroundTable(t, g)
}

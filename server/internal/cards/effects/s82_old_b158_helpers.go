package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b158_helpers.go — shared bodies for Core Set 2020, batch 158 (#5518).

// OldB158SpellTargetsYou is the clause "target spell that targets you": the
// candidate is a spell on the stack with at least one target that is the player
// choosing, as it stands now (CR 115.1). A permanent the player controls does
// not count, and neither does an ability.
func OldB158SpellTargetsYou(g *game.Game, chooser uuid.UUID, c game.Card) bool {
	item := g.StackItemForEffect(c.InstanceID)
	if item == nil || item.Kind != game.StackItemSpell {
		return false
	}
	for _, t := range item.Targets {
		if t.Kind == game.TargetPlayer && t.ID == chooser {
			return true
		}
	}
	return false
}

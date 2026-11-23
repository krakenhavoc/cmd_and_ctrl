package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deathbloom Ritualist — Creature — Elf Warlock {3}{B}{G}, 3/5
// (The Brothers' War):
//
//	"{T}: Add X mana of any one color, where X is the number of creature
//	 cards in your graveyard."
//
// Arbor Adherent's one-colour mana ability (ProducedOneColor) with X read
// from the graveyard when the ability is activated (CR 106.4, 605.3); with
// no creature card there it adds nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b52148c3-fe97-4cc0-a814-a5a0f5d426f6",
		Name:         "Deathbloom Ritualist",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedOneColor(deathbloomRitualistX),
			Label:          "Add X mana of any one color, where X is the number of creature cards in your graveyard",
		}},
	})
}

// deathbloomRitualistX is the number of creature cards in the controller's graveyard.
func deathbloomRitualistX(g *game.Game, controller, _ uuid.UUID) int {
	return effects.B11CreatureCardsInGraveyard(g, controller)
}

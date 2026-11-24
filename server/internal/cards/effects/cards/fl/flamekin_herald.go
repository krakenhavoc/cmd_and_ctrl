package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flamekin Herald — Creature — Elemental Wizard {2}{R}, 3/2 (Commander
// Legends):
//
//	"Commander spells you cast have cascade. (Whenever you cast a commander,
//	 exile cards from the top of your library until you exile a nonland card
//	 with lesser mana value. You may cast it without paying its mana cost. Put
//	 the exiled cards on the bottom in a random order.)"
//
// Maelstrom Nexus's grant (GrantsCascade): a battlefield trigger on the
// controller's cast whose condition is that the spell is a commander, from
// any zone it can be cast from. The cascade limit is that spell's mana value
// as it is cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b3d7411-7a25-4353-83c0-d09525a4fa08",
		Name:         "Flamekin Herald",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GrantsCascade("Flamekin Herald — cascade",
				func(spell game.Card, _ *game.Card, _ *game.Game) bool { return spell.IsCommander }),
		},
	})
}

package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disciple of the Sun — Creature — Human Cleric {4}{W}, 3/3 (Modern
// Horizons 2, #5501):
//
//	"Lifelink
//	 When this creature enters, return target permanent card with mana
//	 value 3 or less from your graveyard to your hand."
//
// Shepherd of the Clouds' target clause, always to hand: picked as the
// trigger goes on the stack, re-checked as it resolves (CR 608.2b), and
// removed for want of a target when there is none (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3ea41848-7a11-4608-88c1-7a6e76e812ca",
		Name:            "Disciple of the Sun",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.WhenThisEnters("Disciple of the Sun — return target permanent card with mana value 3 or less from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
				}),
			effects.TargetCardInGraveyard("target permanent card with mana value 3 or less from your graveyard",
				effects.YouOwn(), effects.Permanent(), effects.ManaValueLE(3)),
		)},
	})
}

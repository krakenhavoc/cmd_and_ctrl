package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghitu Chronicler — Creature — Human Wizard {1}{R}, 1/3 (Dominaria, #5530):
//
//	"Kicker {3}{R} (You may pay an additional {3}{R} as you cast this spell.)
//	 When this creature enters, if it was kicked, return target instant or sorcery card from your graveyard to your hand."
//
// Phyrexian Missionary's kicked return, for an instant or sorcery card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "ab8a00c2-dbb0-48c3-9d13-ad6f1e2f3a18",
		Name:          "Ghitu Chronicler",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{3}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Ghitu Chronicler — kicked: return target instant or sorcery card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
				}),
				effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard", effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery()))),
		},
	})
}

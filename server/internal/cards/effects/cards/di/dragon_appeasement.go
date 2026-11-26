package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon Appeasement — Enchantment {3}{B}{R}{G} (Alara Reborn, #5586):
//
//	"Skip your draw step.
//	 Whenever you sacrifice a creature, you may draw a card."
//
// The draw step is skipped (Yawgmoth's Bargain's replacement). The trigger
// watches the sacrifice event, which fires while the creature is still on the
// battlefield, so a creature is read as it was; the draw is optional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3cb9b08-16d2-4574-afb5-068239dbd152",
		Name:         "Dragon Appeasement",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SkipYourDrawStep()},
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(
				effects.On(game.EventSacrifice, effects.S82OldB501YouSacrificedACreature,
					"Dragon Appeasement — you may draw a card", effects.Do(effects.DrawCards{N: 1})),
				"Dragon Appeasement — draw a card?", 1),
		},
	})
}

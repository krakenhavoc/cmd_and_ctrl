package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Birthing Ritual — Enchantment {1}{G} (Modern Horizons 3):
//
//	"At the beginning of your end step, if you control a creature, look at
//	 the top seven cards of your library. Then you may sacrifice a creature.
//	 If you do, you may put a creature card with mana value X or less from
//	 among those cards onto the battlefield, where X is 1 plus the
//	 sacrificed creature's mana value. Put the rest on the bottom of your
//	 library in a random order."
//
// An intervening "if" (CR 603.4): a creature is checked as the end step
// begins and again as the trigger resolves. The look is a LOOK (only the
// controller sees the seven); the sacrifice is asked as it resolves, with the
// seven carried as plain data. "If you do" is the sacrifice's keyed tail:
// X comes from the sacrificed creature as it last existed (a token that is no
// copy is 0, so X is 1), the put is Armored Skyhunter's, and the rest go to
// the bottom in a random order after it, or at once when the sacrifice was
// declined (gapB55BirthingRitual*, s82_gap_b55_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37a5ab64-12e4-4aad-9a09-46a26fe9953b",
		Name:         "Birthing Ritual",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control a creature",
				Test: func(ctx *effects.Context) bool {
					return len(effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.Creature())) > 0
				},
			}, effects.AtYourEndStep("Birthing Ritual — look at the top seven cards of your library; you may sacrifice a creature",
				effects.GapB55BirthingRitualLook)),
		},
	})
}

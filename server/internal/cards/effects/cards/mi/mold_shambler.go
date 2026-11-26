package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mold Shambler — Creature — Fungus Beast {3}{G}, 3/3 (Commander 2013, S82
// older sets):
//
//	"Kicker {1}{G} (You may pay an additional {1}{G} as you cast this spell.)
//	 When this creature enters, if it was kicked, destroy target noncreature
//	 permanent."
//
// Kor Sanctifiers' shape: an enters trigger that exists only when the spell
// was kicked (CR 702.33d, 603.4), with its target chosen as the trigger goes
// on the stack and judged again as it resolves (CR 608.2b). A noncreature
// permanent is any permanent that is not a creature, lands included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "7aba1039-f108-423e-8aa4-d9d8b4da2414",
		Name:          "Mold Shambler",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Mold Shambler — kicked, destroy target noncreature permanent",
					effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target noncreature permanent", effects.Not(effects.Creature()))),
		},
	})
}

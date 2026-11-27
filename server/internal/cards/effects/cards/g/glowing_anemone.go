package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glowing Anemone — Creature — Jellyfish Beast {3}{U}, 1/3 (Mercadian Masques,
// S82 OLD batch 700, #5619):
//
//	"When this creature enters, you may return target land to its owner's
//	 hand."
//
// Aven Fogbringer's bounce, made optional like Moriok Scavenger's: it targets a
// land of any player's, so with no land there is no trigger (CR 603.3d), and
// the "may" is asked as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52305f9d-2bfd-4b35-92b3-c08cdd758cde",
		Name:         "Glowing Anemone",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Glowing Anemone — you may return target land to its owner's hand", effects.BounceTheTargetToHand),
				effects.TargetPermanent("target land", effects.Land())),
				"Glowing Anemone — return the land to its owner's hand?"),
		},
	})
}

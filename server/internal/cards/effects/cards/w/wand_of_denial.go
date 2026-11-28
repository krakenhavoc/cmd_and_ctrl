package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wand of Denial — Artifact {2} (Visions, S82 older sets batch 791, #5631):
//
//	"{T}: Look at the top card of target player's library. If it's a nonland
//	 card, you may pay 2 life. If you do, put it into that player's graveyard."
//
// Only the controller sees the card (CR 701.20 is not a reveal). For a nonland
// card they are offered the 2 life; paying it puts that card, if it is still on
// top of that library, into its owner's graveyard. A controller who can't pay 2
// life is not asked (CR 118.3), and a target that is no longer legal does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d32da71-60e5-4eb3-97ca-ae76e2f5df57",
		Name:         "Wand of Denial",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Look at the top card of target player's library. If it's a nonland card, you may pay 2 life. If you do, put it into that player's graveyard.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.Old791WandOfDenial,
		}},
	})
}

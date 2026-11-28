package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undying Beast — Creature — Beast {3}{B}, 3/2 (Portal, S82 OLD batch 783,
// #5630):
//
//	"When this creature dies, put it on top of its owner's library."
//
// Dark Revenant's dies trigger: the card moves only while it is still the same
// card in its owner's graveyard (CR 400.7), through the tuck route, so a
// commander's owner is offered the command zone (CR 903.9) and a token, which
// ceased to exist, is not found.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ae03181-a436-4bad-be3a-6c9f6c0ed4d6",
		Name:         "Undying Beast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Undying Beast — put it on top of its owner's library", effects.PutThisDiedCardOnTopOfItsOwnersLibrary),
		},
	})
}

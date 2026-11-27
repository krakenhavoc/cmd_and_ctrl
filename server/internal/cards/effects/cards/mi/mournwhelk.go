package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mournwhelk — Creature — Elemental {6}{B}, 3/3 (Lorwyn, S82 OLD, #5593):
//
//	"When this creature enters, target player discards two cards.
//	 Evoke {3}{B}"
//
// Mulldrifter's shape: the enters trigger is real and so is the evoke sacrifice
// (CR 702.74a), so an evoked Mournwhelk still makes its target discard two cards
// before it is sacrificed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "10277cc5-f53b-4b33-bd68-edb0534a2040",
		Name:             "Mournwhelk",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Evoke("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Mournwhelk — target player discards two cards", effects.TargetPlayerDiscards(2)),
				effects.TargetPlayer("target player")),
		},
	})
}

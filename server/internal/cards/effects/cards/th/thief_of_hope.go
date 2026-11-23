package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thief of Hope — Creature — Spirit {2}{B}, 2/2 (Champions of Kamigawa,
// #5603):
//
//	"Whenever you cast a Spirit or Arcane spell, target opponent loses 1
//	 life and you gain 1 life.
//	 Soulshift 2"
//
// Blood Artist's drain on the ki cycle's cast trigger: the opponent is
// chosen as it goes on the stack, and with that opponent gone the whole
// ability does nothing, the life gain included (CR 608.2b). Soulshift 2
// is the shared keyword trigger (CR 702.46a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3fa1320-9e95-437a-aa36-102a2722fc42",
		Name:         "Thief of Hope",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverYouCast(effects.ChkB1SpiritOrArcane(), "Thief of Hope — target opponent loses 1 life and you gain 1 life",
				effects.Do(effects.TargetPlayerLosesLife{N: 1}, effects.GainLife{Amount: 1})),
				effects.TargetPlayer("target opponent", effects.Opponent())),
			effects.Soulshift(2, "Thief of Hope"),
		},
	})
}

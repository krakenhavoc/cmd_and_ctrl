package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Diligent Excavator — Creature — Human Artificer {1}{U}, 1/3 (Dominaria, S82
// OLD batch 215, #5530):
//
//	"Whenever you cast a historic spell, target player mills two cards.
//	 (Artifacts, legendaries, and Sagas are historic.)"
//
// Jamie McCrimmon's cast trigger with a player target chosen as it goes on the
// stack, re-checked as it resolves (CR 608.2b); the player may be you. Only
// your own casts count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3866181-c127-45d6-8553-97e9f64a3ca5",
		Name:         "Diligent Excavator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverYouCast(effects.B14Historic(), "Diligent Excavator — target player mills two cards", effects.TargetPlayerMillsEffect(2)),
				effects.TargetPlayer("target player")),
		},
	})
}

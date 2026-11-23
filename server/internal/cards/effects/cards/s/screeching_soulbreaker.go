package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Screeching Soulbreaker — Creature — Siren Bard {2}{B}, 1/4:
//
//	"Flying
//	 Whenever this creature attacks, it deals 1 damage to each opponent
//	 and you gain 1 life."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "47e33aa1-2f54-471c-a484-de13bcd2dceb",
		Name:            "Screeching Soulbreaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Screeching Soulbreaker — 1 damage to each opponent, you gain 1 life",
				effects.DamageEachOpponentThenGainLife(1)),
		},
	})
}

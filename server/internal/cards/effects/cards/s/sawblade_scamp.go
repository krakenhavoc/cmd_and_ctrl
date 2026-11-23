package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sawblade Scamp — Creature — Phyrexian Beast {R}, 1/1 (Phyrexia: All Will Be One):
//
//	"Haste
//	 Whenever you cast a noncreature spell, put an oil counter on this creature.
//	 {T}, Remove an oil counter from this creature: It deals 1 damage to each opponent."
//
// The tap and the counter are both paid at announce, so a Scamp with no oil
// counter cannot be activated and a summoning-sick one cannot either (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8cac286c-1a17-4ee1-9858-d7a2f8c09037",
		Name:            "Sawblade Scamp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{effects.S82OneB09OilOnNoncreatureCast("Sawblade Scamp")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove an oil counter from this creature: It deals 1 damage to each opponent.",
			Purpose: game.Purpose{DamageEachOpponent: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis(effects.S82OneB09Oil, 1)),
			Effect:  effects.Do(effects.DamageEachOpponent{N: 1}),
		}},
	})
}

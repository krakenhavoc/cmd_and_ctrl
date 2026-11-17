package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "8cac286c-1a17-4ee1-9858-d7a2f8c09037",
		Name:            "Sawblade Scamp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{s82OneB09OilOnNoncreatureCast("Sawblade Scamp")},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Remove an oil counter from this creature: It deals 1 damage to each opponent.",
			Purpose: game.Purpose{DamageEachOpponent: 1, Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), RemoveCountersFromThis(s82OneB09Oil, 1)),
			Effect:  Do(DamageEachOpponent{N: 1}),
		}},
	})
}

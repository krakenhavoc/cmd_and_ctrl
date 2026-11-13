package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dedicated Martyr — Creature — Human Cleric {W}, 1/1 (Odyssey, #5613):
//
//	"{W}, Sacrifice this creature: You gain 3 life."
//
// Ice Cream Kitty's life row with {W} for {2}, {T}: the sacrifice is part
// of the cost (CR 602.2b), so the creature is gone before the ability
// resolves and the life comes anyway. Without a tap symbol it works the
// turn the Martyr arrives. It answers nothing (ADR 0142).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca39af65-985d-4ab1-874b-9e8d3455c349",
		Name:         "Dedicated Martyr",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, Sacrifice this creature: You gain 3 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue, LifeGain: 3},
			Cost:    Plus(ManaCost("{W}"), SacrificeThis()),
			Effect:  Do(GainLife{Amount: 3}),
		}},
	})
}

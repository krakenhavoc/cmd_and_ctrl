package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frostburn Weird — Creature — Weird {U/R}{U/R}, 1/4 (Ravnica: Clue
// Edition, #5929):
//
//	"{U/R}: This creature gets +1/-1 until end of turn."
//
// A hybrid-mana pump, no tap in the cost, so it can be activated any
// number of times. The toughness drop is a real -1: at 0 toughness the
// Weird dies to the state-based action (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f667b4b6-c97f-47fa-9146-c5c1f3f47e92",
		Name:         "Frostburn Weird",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U/R}: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
			Cost:    ManaCost("{U/R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{Target: ctx.Source(), Power: 1, Toughness: -1, Label: "Frostburn Weird — +1/-1"}.Apply(ctx)
			},
		}},
	})
}

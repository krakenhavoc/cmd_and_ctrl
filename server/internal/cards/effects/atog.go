package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Atog — Creature — Atog {1}{R}, 1/2 (Antiquities, #5640):
//
//	"Sacrifice an artifact: This creature gets +2/+2 until end of turn."
//
// A free activation whose only cost is one artifact the controller picks
// at announce. Each activation is its own +2/+2 until end of turn, so
// the Atog grows by two for every artifact it eats.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3095f2cf-05c3-4381-929f-c526b6fc30a7",
		Name:         "Atog",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice an artifact: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    SacrificeN(1, "an artifact", Artifact()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return (BoostUntilEOT{Target: item.SourceCardID, Power: 2, Toughness: 2, Label: "Atog — +2/+2"}).Apply(NewContext(g, item))
			},
		}},
	})
}

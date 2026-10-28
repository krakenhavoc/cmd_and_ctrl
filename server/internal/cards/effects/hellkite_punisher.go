package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hellkite Punisher — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "c3ea33f4-17ac-41a4-8320-361b6f8aba0d", Name: "Hellkite Punisher", Completeness: CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{Label: "{R}: This creature gets +1/+0 until end of turn.", Cost: ManaCost("{R}"), Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return (BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Hellkite Punisher — +1/+0"}).Apply(NewContext(g, item))
			},
		}},
	})
}

package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hellkite Punisher — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "c3ea33f4-17ac-41a4-8320-361b6f8aba0d", Name: "Hellkite Punisher", Completeness: effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{Label: "{R}: This creature gets +1/+0 until end of turn.", Cost: effects.ManaCost("{R}"), Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Hellkite Punisher — +1/+0"}).Apply(effects.NewContext(g, item))
			},
		}},
	})
}

package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thorn-Thrash Viashino — Creature — Lizard Warrior {3}{R}, 2/2 (Planechase
// 2012, S82 OLD, #5569):
//
//	"Devour 2 (As this creature enters, you may sacrifice any number of
//	 creatures. It enters with twice that many +1/+1 counters on it.)
//	 {G}: This creature gains trample until end of turn."
//
// Devour 2 is Predator Dragon's (CR 702.82a). The activated ability is a
// keyword grant to the creature itself until end of turn (CR 611.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83d9a5a2-09b0-44bf-b0b8-82a5fc53b00a",
		Name:         "Thorn-Thrash Viashino",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Devour("Thorn-Thrash Viashino", 2)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}: This creature gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"trample"},
					Label: "Thorn-Thrash Viashino — trample until end of turn"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

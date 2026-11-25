package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pitiless Pontiff — Creature — Vampire Cleric {W}{B}, 2/2 (Ravnica
// Allegiance):
//
//	"{1}, Sacrifice another creature: This creature gains deathtouch and
//	 indestructible until end of turn."
//
// The sacrifice is a cost of "another" creature (never the Pontiff); both
// keywords are granted to the Pontiff itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46a9f5af-b024-497d-9f55-98d17e2ffa76",
		Name:         "Pitiless Pontiff",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice another creature: This creature gains deathtouch and indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"deathtouch", "indestructible"},
					Label: "Pitiless Pontiff — deathtouch and indestructible until end of turn"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

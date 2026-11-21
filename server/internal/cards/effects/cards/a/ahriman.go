package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ahriman — Creature — Eye Horror {2}{B}, 2/2 (Final Fantasy, #4462):
//
//	"Flying, deathtouch
//	 {3}, Sacrifice another creature or artifact: Draw a card."
//
// The sacrifice is part of the cost (CR 602.2), so the creature or
// artifact is gone before the ability can be answered, and "another"
// keeps Ahriman from paying with itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "37d1d507-25dc-4cd0-b51e-2f1e9b24832c",
		Name:            "Ahriman",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, Sacrifice another creature or artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: effects.Plus(effects.ManaCost("{3}"), game.AbilityCost{
				SacrificeOther: effects.Another(effects.SacrificeSpec("another creature or artifact",
					effects.Or(effects.Creature(), effects.Artifact()))),
			}),
			Effect: effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}

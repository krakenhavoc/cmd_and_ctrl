package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akki Scrapchomper — Creature — Phyrexian Goblin 1/1 {R} (March of the Machine, S82):
//
//	"{1}{R}, {T}, Sacrifice an artifact or land: Draw a card."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "baea7cc5-15ec-4229-ad8b-d4b34c2e21e8",
		Name:         "Akki Scrapchomper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, {T}, Sacrifice an artifact or land: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: effects.Plus(effects.ManaCost("{1}{R}"), game.AbilityCost{
				Tap:            true,
				SacrificeOther: effects.SacrificeSpec("an artifact or land", effects.Or(effects.Artifact(), effects.Land())),
			}),
			Effect: effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}

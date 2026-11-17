package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Akki Scrapchomper — Creature — Phyrexian Goblin 1/1 {R} (March of the Machine, S82):
//
//	"{1}{R}, {T}, Sacrifice an artifact or land: Draw a card."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "baea7cc5-15ec-4229-ad8b-d4b34c2e21e8",
		Name:         "Akki Scrapchomper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}, {T}, Sacrifice an artifact or land: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: Plus(ManaCost("{1}{R}"), game.AbilityCost{
				Tap:            true,
				SacrificeOther: sacrificeSpec("an artifact or land", Or(Artifact(), Land())),
			}),
			Effect: Do(DrawCards{N: 1}),
		}},
	})
}

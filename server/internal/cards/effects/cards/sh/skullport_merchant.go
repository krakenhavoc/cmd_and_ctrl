package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skullport Merchant — Creature — Dwarf Citizen {2}{B}, 1/4 (AFR):
//
//	"When this creature enters, create a Treasure token.
//	 {1}{B}, Sacrifice another creature or a Treasure: Draw a card."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64ed5c35-8185-4e68-8028-b7b632ef5828",
		Name:         "Skullport Merchant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Skullport Merchant — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice another creature or a Treasure: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerSacOutlet},
			Cost: game.AbilityCost{
				Mana: "{1}{B}",
				SacrificeOther: effects.Another(effects.SacrificeSpec("another creature or a Treasure",
					effects.Or(effects.Creature(), effects.HasSubtype("Treasure")))),
			},
			Effect: effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}

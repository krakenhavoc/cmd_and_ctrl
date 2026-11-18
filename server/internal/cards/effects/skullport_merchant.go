package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skullport Merchant — Creature — Dwarf Citizen {2}{B}, 1/4 (AFR):
//
//	"When this creature enters, create a Treasure token.
//	 {1}{B}, Sacrifice another creature or a Treasure: Draw a card."
func init() {
	Register(Spec{
		OracleID:     "64ed5c35-8185-4e68-8028-b7b632ef5828",
		Name:         "Skullport Merchant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Skullport Merchant — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice another creature or a Treasure: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerSacOutlet},
			Cost: game.AbilityCost{
				Mana: "{1}{B}",
				SacrificeOther: Another(sacrificeSpec("another creature or a Treasure",
					Or(Creature(), HasSubtype("Treasure")))),
			},
			Effect: Do(DrawCards{N: 1}),
		}},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spear Spewer — Creature — Goblin Warrior {R}, 1/1 (Jumpstart 2022, #5948):
//
//	"Defender
//	 {T}: This creature deals 1 damage to each player."
//
// Ashcloud Phoenix's damage to each player, from the Spewer: one damage
// instance, so a prevention or redirect sees each player's share.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1a668cef-bdc5-410b-b9e3-044e4203a0aa",
		Name:            "Spear Spewer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to each player.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b38DamageEachPlayer(NewContext(g, item), 1)
			},
		}},
	})
}

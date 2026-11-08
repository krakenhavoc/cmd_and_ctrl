package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teapot Slinger — Creature — Raccoon Warrior {3}{R}, 3/4 (Bloomburrow,
// #4467):
//
//	"Menace
//	 Whenever you expend 4, this creature deals 2 damage to each opponent.
//	 (You expend 4 as you spend your fourth total mana to cast spells
//	 during a turn.)"
//
// Expend is WheneverYouExpend (expend.go), which fires once a turn, as
// the fourth mana is spent. The damage is Angry Rabble's, dealt by the
// Slinger to every opponent still in the game.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5ba19dbe-9dbb-4cc9-81a3-18c26f27bd0f",
		Name:            "Teapot Slinger",
		ReadsManaSpent:  game.ReadsYourCasts,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverYouExpend(4, "Teapot Slinger — 2 damage to each opponent", Do(DamageEachOpponent{N: 2})),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Luke Cage, Hero for Hire — Legendary Creature — Human Mercenary Hero
// {2}{R}{R}, 4/4 (Marvel Super Heroes Commander, #4455):
//
//	"Trample
//	 At the beginning of combat on your turn, create a Treasure token."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3d776a25-c341-4058-9b5f-84e70edb25a4",
		Name:            "Luke Cage, Hero for Hire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Luke Cage — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}

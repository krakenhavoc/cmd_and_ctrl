package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mesa Cavalier — Creature — Human Knight {2}{W}, 2/1 (Dominaria United, S82):
//
//	"Flying
//	 When this creature enters, you gain 2 life."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "898763f8-a333-43f5-9f66-52e2adbe5758",
		Name:            "Mesa Cavalier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Mesa Cavalier — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}

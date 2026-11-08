package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vinereap Mentor — Creature — Squirrel Druid {B}{G}, 3/2 (Bloomburrow,
// #4467):
//
//	"When this creature enters or dies, create a Food token. (It's an
//	 artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")"
//
// Crow of Dark Tidings' one ability with two conditions
// (WhenThisEntersOrDies), making the catalog's printed Food token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "db649e15-24ca-4fd8-9f0b-f1df28c9c57d",
		Name:         "Vinereap Mentor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrDies("Vinereap Mentor — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
	})
}

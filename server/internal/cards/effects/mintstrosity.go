package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mintstrosity — Creature — Horror {1}{B}, 3/1 (Wilds of Eldraine, #4471):
//
//	"When this creature dies, create a Food token. (It's an artifact with
//	 "{2}, {T}, Sacrifice this token: You gain 3 life.")"
//
// A dies trigger whose payout is the catalog's Food token, made under the
// trigger's controller: the player who controlled the Horror as it died.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8c88c33-b8df-4b19-b12c-79c9709af8f1",
		Name:         "Mintstrosity",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Mintstrosity — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tolarian Serpent — Creature — Serpent {5}{U}{U}, 7/7 (Weatherlight, #5629):
//
//	"At the beginning of your upkeep, mill seven cards."
//
// Nyx Weaver's upkeep mill, seven deep (CR 701.17a). A library holding
// fewer mills what it has.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "684e18bf-1390-4d12-8e7a-e8562db64dbb",
		Name:         "Tolarian Serpent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Tolarian Serpent — mill seven cards", Do(MillCards{N: 7})),
		},
	})
}

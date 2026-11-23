package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tolarian Serpent — Creature — Serpent {5}{U}{U}, 7/7 (Weatherlight, #5629):
//
//	"At the beginning of your upkeep, mill seven cards."
//
// Nyx Weaver's upkeep mill, seven deep (CR 701.17a). A library holding
// fewer mills what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "684e18bf-1390-4d12-8e7a-e8562db64dbb",
		Name:         "Tolarian Serpent",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Tolarian Serpent — mill seven cards", effects.Do(effects.MillCards{N: 7})),
		},
	})
}

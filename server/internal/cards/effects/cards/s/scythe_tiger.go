package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scythe Tiger — Creature — Cat {G}, 3/2 (Zendikar, S82 OLD b494):
//
//	"Shroud
//	 When this creature enters, sacrifice it unless you sacrifice a land."
//
// The Gitrog Monster's upkeep body (SacrificeThisUnlessYouSacrificeALand) on an
// enters trigger: the controller sacrifices a land, or the Tiger; with no land
// to sacrifice the Tiger goes. Shroud is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4d0fa514-87fc-46ff-bf49-4cd332396978",
		Name:            "Scythe Tiger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shroud"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Scythe Tiger — sacrifice it unless you sacrifice a land", effects.SacrificeThisUnlessYouSacrificeALand("Scythe Tiger")),
		},
	})
}

package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bog Elemental — Creature — Elemental {3}{B}{B}, 5/4 (Prophecy, S82 OLD batch
// 673, #5617):
//
//	"Protection from white
//	 At the beginning of your upkeep, sacrifice this creature unless you
//	 sacrifice a land."
//
// Protection from white is a printed keyword. The upkeep body is The Gitrog
// Monster's (SacrificeThisUnlessYouSacrificeALand): the controller sacrifices a
// land of their choice, or the Elemental; with no land to sacrifice it goes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "82344b5b-7cd9-460d-9966-f09432ed8e59",
		Name:            "Bog Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from white"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Bog Elemental — sacrifice it unless you sacrifice a land", effects.SacrificeThisUnlessYouSacrificeALand("Bog Elemental")),
		},
	})
}

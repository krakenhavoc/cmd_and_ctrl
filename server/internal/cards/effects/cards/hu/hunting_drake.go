package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hunting Drake — Creature — Drake {4}{U}, 2/2 (Planeshift, S82 OLD b643, #5615):
//
//	"Flying
//	 When this creature enters, put target red or green creature on top of its
//	 owner's library."
//
// Flying is the printed keyword. The trigger targets a red or green creature (any
// controller's, the Drake itself is neither) and tucks it on top of its owner's
// library as the trigger resolves (CR 608.2b: a target that has become illegal is
// not moved).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d5978943-52fc-4b57-9c1c-d802ade1b014",
		Name:            "Hunting Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Hunting Drake — put target red or green creature on top of its owner's library", effects.OldB643TuckFirstTarget),
				effects.TargetCreature("target red or green creature", effects.Or(effects.OfColor("R"), effects.OfColor("G")))),
		},
	})
}

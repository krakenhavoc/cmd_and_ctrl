package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icefeather Aven — Creature — Bird Shaman {G}{U}, 2/2 (Commander 2019):
//
//	"Flying
//	 Morph {1}{G}{U} (You may cast this card face down as a 2/2 creature for
//	 {3}. Turn it face up any time for its morph cost.)
//	 When this creature is turned face up, you may return another target
//	 creature to its owner's hand."
//
// Morph is the shared constructor. The face-up trigger is Midnight Scavengers'
// optional targeted trigger: the target (any creature but this one) is chosen as
// the ability goes on the stack, and the "you may" is asked as it resolves
// (CR 603.5); a target that has left does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "9fb901db-1ff7-4a8d-8233-fe7e3dee7fd1",
		Name:             "Icefeather Aven",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{1}{G}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisIsTurnedFaceUp("Icefeather Aven — you may return another target creature to its owner's hand", effects.BounceChosenTarget),
				effects.Another(effects.TargetCreature("another target creature"))),
				"Icefeather Aven — return the creature to its owner's hand?"),
		},
	})
}

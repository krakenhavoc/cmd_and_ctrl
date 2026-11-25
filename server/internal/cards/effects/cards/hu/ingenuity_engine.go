package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ingenuity Engine — Artifact {7} (Commander Legends):
//
//	"Cascade (When you cast this spell, exile cards from the top of your
//	 library until you exile a nonland card that costs less. You may cast it
//	 without paying its mana cost. Put the exiled cards on the bottom in a
//	 random order.)
//	 {1}, {T}, Sacrifice an artifact: Return target artifact you control to
//	 its owner's hand."
//
// The engine's cascade trigger, and an activated ability whose cost
// sacrifices an artifact (the Engine itself is a legal choice) and whose
// target is chosen first, so the sacrificed artifact may be the target, in
// which case the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1da2c33-2be4-4065-8e5c-613a069d2013",
		Name:         "Ingenuity Engine",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Cascade()},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}, {T}, Sacrifice an artifact: Return target artifact you control to its owner's hand.",
			Cost: effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), game.AbilityCost{
				SacrificeOther: effects.SacrificeSpec("an artifact", effects.Artifact()),
			}),
			Targets: effects.TargetPermanent("target artifact you control", effects.Artifact(), effects.YouControl()),
			Effect:  effects.BounceEachLegalTarget,
		}},
	})
}

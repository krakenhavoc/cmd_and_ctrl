package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Atarka Efreet — Creature — Efreet Shaman {3}{R}, 5/1 (Dragons of Tarkir,
// #5551):
//
//	"Megamorph {2}{R} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its megamorph cost and put a +1/+1 counter on it.)
//	 When this creature is turned face up, it deals 1 damage to any target."
//
// Megamorph is the shared constructor (morph.go): the face-down cast for
// {3}, and the +1/+1 counter placed by the turn-face-up special action
// (CR 702.37b).
// The face-up trigger is Skeleton Archer's enters trigger on the
// turned-face-up event: the target is chosen as it goes on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "177c1f2d-abb9-40ea-afe4-141b08fdeabb",
		Name:             "Atarka Efreet",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Megamorph("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisIsTurnedFaceUp("Atarka Efreet — 1 damage to any target",
				sourceDealsDamageToEachLegalTarget(1)), TargetAny()), ForTargets(DamageToTarget(0, 1))),
		},
	})
}

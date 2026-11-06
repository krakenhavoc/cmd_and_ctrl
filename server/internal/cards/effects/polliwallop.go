package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Polliwallop — Instant {3}{G} (Bloomburrow, #4467):
//
//	"Affinity for Frogs (This spell costs {1} less to cast for each Frog
//	 you control.)
//	 Target creature you control deals damage equal to twice its power to
//	 target creature you don't control."
//
// Affinity (CR 702.41a) is the shared AffinityFor cost modifier, counted
// when the total cost is determined (CR 601.2f) and taken off the generic
// part only (CR 118.7a), so the spell never costs less than {G}. The bite
// is Punishing Punch's resolution (slot 0 the damage source, slot 1 the
// recipient, the power read as the spell resolves and doubled); only the
// second clause differs, "you don't control" rather than "an opponent
// controls".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "efd2a4ca-1bd0-4762-a997-f099326b1746",
		Name:         "Polliwallop",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", Not(YouControl())),
		),
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for Frogs", HasSubtype("Frog")),
		},
		OnResolve: ownCreatureDealsTwiceItsPowerToOpposingCreature,
	})
}

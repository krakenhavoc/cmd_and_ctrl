package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Amanda Grayson, Logic's Link — Legendary Creature — Human Advisor
// {3}{G}, 2/3 (Star Trek Commander, #4974):
//
//	"Affinity for Vulcans (This spell costs {1} less to cast for each
//	 Vulcan you control.)
//	 Landfall — Whenever a land you control enters, put two +1/+1
//	 counters on target creature you control."
//
// Affinity is Cantankerous Keepers' self cost modifier over Vulcans you
// control (CR 702.41a), read through the layers so a changeling counts.
// The landfall trigger targets a creature you control and puts the two
// counters on it if it is still a legal target as the trigger resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	landfall := Landfall("Amanda Grayson, Logic's Link — put two +1/+1 counters on target creature you control",
		rfCreatureFCountersOnTarget(2))
	landfall.Targets = TargetCreature("target creature you control", YouControl())
	Register(Spec{
		OracleID:     "13d4a70f-c48c-4739-8f3e-5f0302c5484c",
		Name:         "Amanda Grayson, Logic's Link",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for Vulcans", HasSubtype("Vulcan")),
		},
		Triggered: []game.TriggeredAbility{landfall},
	})
}

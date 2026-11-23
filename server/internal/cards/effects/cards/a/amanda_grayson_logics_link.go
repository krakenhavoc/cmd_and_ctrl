package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	landfall := effects.Landfall("Amanda Grayson, Logic's Link — put two +1/+1 counters on target creature you control",
		effects.RfCreatureFCountersOnTarget(2))
	landfall.Targets = effects.TargetCreature("target creature you control", effects.YouControl())
	effects.Register(effects.Spec{
		OracleID:     "13d4a70f-c48c-4739-8f3e-5f0302c5484c",
		Name:         "Amanda Grayson, Logic's Link",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for Vulcans", effects.HasSubtype("Vulcan")),
		},
		Triggered: []game.TriggeredAbility{landfall},
	})
}

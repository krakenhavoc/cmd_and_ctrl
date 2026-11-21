package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Explosive Singularity — Sorcery {8}{R}{R} (#3961):
//
//	"As an additional cost to cast this spell, you may tap any number of
//	 untapped creatures you control. This spell costs {1} less to cast
//	 for each creature tapped this way.
//	 Explosive Singularity deals 10 damage to any target."
//
// The additional cost is TapAnyNumberUntappedCost (Burn at the Stake's):
// the creatures ride cast_spell as teamwork_ids, none included, and are
// tapped with the spell on the stack (CR 601.2h). Their number is
// announced before the total cost is settled (CR 601.2b before 601.2f),
// so CostsLessPerTappedToPay takes {1} of generic mana off per creature
// through the one pricer: the price CastSpell charges, the auto-tap
// preview shows and the enumerator offers. A reduction never touches the
// {R}{R}, so eight creatures or more cast it for {R}{R}. A creature
// tapped for the cost cannot also tap for mana: the auto-tapper keeps the
// named creatures out of the mana it plans.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "30d79f61-f9d9-441a-b600-83721af8553b",
		Name:           "Explosive Singularity",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.TapAnyNumberUntappedCost("any number of untapped creatures you control", effects.Creature()),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessPerTappedToPay("{1}", "This spell costs {1} less to cast for each creature tapped this way"),
		},
		Targets:   effects.TargetAny(),
		Purpose:   effects.ForTargets(effects.DamageToTarget(0, 10)),
		OnResolve: effects.DamageToFirstTargetStep(10),
	})
}

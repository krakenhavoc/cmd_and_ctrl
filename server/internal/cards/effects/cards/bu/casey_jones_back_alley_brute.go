package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Casey Jones, Back Alley Brute — {3}{R} Legendary Creature — Human
// Berserker 4/2 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Menace
//	 Whenever Casey Jones attacks, put a +1/+1 counter on target
//	 attacking creature.
//	 Whenever you put one or more +1/+1 counters on a creature you
//	 control, Casey Jones deals that much damage to target opponent."
//
// The attack trigger targets any attacking creature, Casey Jones
// included (Araña's clause). The second ability is Terrasymbiosis's
// "whenever you put": once per creature per batch of +1/+1 counters its
// controller put (an opponent's effect growing your creature is not
// you putting them), and "that much" is the number of +1/+1 counters put
// on that creature in the batch. So the attack trigger's own counter
// deals 1. The damage is Casey Jones's, from its last-known information
// if it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ab9ff8e9-0ffa-4ccf-a4bf-8505079616f6",
		Name:            "Casey Jones, Back Alley Brute",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Casey Jones — put a +1/+1 counter on target attacking creature",
					effects.Do(effects.Instruction(effects.CounterOnFirstCreatureTarget))),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature())),
			effects.Targeting(
				effects.WheneverYouPutCountersOfKindOnACreature(game.CounterPlusOne, effects.CreatureYouControl,
					"Casey Jones — deals that much damage to target opponent", caseyJonesThatMuchDamage),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

func caseyJonesThatMuchDamage(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	n := effects.CountersOfKindYouPutInTheSameBatch(g, item.Trigger.Event, item.Controller, game.CounterPlusOne)
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
		}
	}
	return nil
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Warrior's Resolve — Enchantment {2}{W} (Final Fantasy, #4462):
//
//	"Creatures you control have training. (Whenever a creature you control
//	 attacks with another creature with greater power, put a +1/+1 counter
//	 on the creature with lesser power.)
//	 At the beginning of your end step, if you control a creature with a
//	 +1/+1 counter on it that attacked this turn, draw a card."
//
// Training is the shared trigger (training.go), granted as a layer 6
// bundle to every creature you control, so it is the creature's own and
// "another creature with greater power" reads that creature (CR 702.149a).
// A creature that loses the enchantment loses it with it.
//
// The end step trigger has a CR 603.4 intervening "if": asked as it
// triggers and again as it resolves, of the creatures you control as they
// are then, so a creature with a counter now counts even if the counter
// came after it attacked (the ruling), and one that attacked this turn is
// read per object (CR 400.7).
//
// No simplification.
const warriorsResolveTrainingGrant = "warriors-resolve/training"

func init() {
	Register(Spec{
		OracleID:     "e3f68aff-9347-452a-8f26-b344016a275d",
		Name:         "Warrior's Resolve",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       warriorsResolveTrainingGrant,
			Triggered: []game.TriggeredAbility{Training("Training — put a +1/+1 counter on this creature")},
			Text:      "Training (Whenever this creature attacks with another creature with greater power, put a +1/+1 counter on this creature.)",
		}},
		Static: []game.StaticAbility{GrantAbilities(CreaturesYouControl, warriorsResolveTrainingGrant)},
		Triggered: []game.TriggeredAbility{
			InterveningIf(warriorsResolveCondition, On(game.EventBeginEndStep,
				AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return controlsACounteredCreatureThatAttacked(g, source.Controller)
				}),
				"Warrior's Resolve — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}

// warriorsResolveCondition is the intervening "if" as the item resolves.
var warriorsResolveCondition = CondBoard{
	Name: "you control a creature with a +1/+1 counter on it that attacked this turn",
	Test: func(ctx *Context) bool { return controlsACounteredCreatureThatAttacked(ctx.Game, ctx.Controller()) },
}

// controlsACounteredCreatureThatAttacked is "you control a creature with a
// +1/+1 counter on it that attacked this turn": the counter as the creature
// is now, the attack as this turn's per-object record.
func controlsACounteredCreatureThatAttacked(g *game.Game, you uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.IsCreature() && c.Counters[game.CounterPlusOne] > 0 && g.AttackedThisTurn(c.InstanceID) {
			return true
		}
	}
	return false
}

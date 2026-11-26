package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sidisi's Faithful — Creature — Snake Wizard {U}, 0/4 (Dragons of Tarkir, #5551):
//
//	"Exploit (When this creature enters, you may sacrifice a creature.)
//	 When this creature exploits a creature, return target creature to its owner's hand."
//
// Exploit is the shared constructor (exploit.go, CR 702.110); the payoff is a reflexive trigger raised only when a creature was really sacrificed.
// The target is chosen when the payoff goes on the stack, and may be any creature.
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "45de1caf-46c2-4acc-8be2-08e1840a3b98",
		Name:         "Sidisi's Faithful",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Exploit("Sidisi's Faithful", sidisisFaithfulBounceBody),
		},
	})
}

// sidisisFaithfulBounceBody is "When this creature exploits a creature, return target creature to its owner's hand." An on-disk key:
// never renamed, never reused.
var sidisisFaithfulBounceBody = game.ReflexiveBody("sidisis-faithful/bounce", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.BounceToHand{Target: id}.Apply(ctx)
}), effects.ConstTargets(func() *game.TargetSpec { return effects.TargetCreature("target creature") }))

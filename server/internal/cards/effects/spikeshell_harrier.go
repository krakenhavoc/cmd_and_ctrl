package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spikeshell Harrier — Artifact Creature — Robot Turtle {4}{U}, 4/4
// (Aetherdrift, #4464):
//
//	"When this creature enters, return target creature or Vehicle an
//	 opponent controls to its owner's hand. If that opponent's speed is
//	 greater than each other player's speed, reduce that opponent's speed
//	 by 1. This effect can't reduce their speed below 1."
//
// The bounce is one sentence and the speed clause the next, in printed order
// (CR 608.2c). "That opponent" is the target's controller as it resolves,
// read before the permanent leaves, and the clause happens whether or not
// the bounce moved the card (a commander sent to the command zone, say).
// The comparison is against every other player still in the game, not just
// opponents, as printed; the speed seam is game.ReduceSpeedForEffect
// (speed_reduce.go), which floors at 1 and goes through the one speed
// write, so a speed that falls from 4 ends max speed. A target gone by
// resolution is illegal and the whole ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23102d8a-3a93-4e83-af5b-401c85ba9dbc",
		Name:         "Spikeshell Harrier",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Spikeshell Harrier — return target creature or Vehicle an opponent controls to its owner's hand; reduce that opponent's speed",
					spikeshellHarrierBounce),
				TargetPermanent("target creature or Vehicle an opponent controls",
					Or(Creature(), HasSubtype("Vehicle")), OpponentControls())),
		},
	})
}

func spikeshellHarrierBounce(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	opp, ok := controllerOfTarget(ctx, id)
	if !ok {
		return nil
	}
	// The speed clause is about the player, not the card, so it runs once
	// the bounce has settled whatever became of the card (a commander's
	// CR 903.9 question included), with the opponent as the tail's player.
	then := ctx.Tail(spikeshellHarrierSpeed)
	then.Player = opp
	return BounceToHand{Target: id, ThenTail: then}.Apply(ctx)
}

// spikeshellHarrierSpeed is "if that opponent's speed is greater than each
// other player's speed, reduce that opponent's speed by 1. This effect
// can't reduce their speed below 1."
var spikeshellHarrierSpeed = CardTail("spikeshell-harrier/reduce-speed", func(ctx *Context, o game.TailOutcome) error {
	g, opp := ctx.Game, o.Tail.Player
	if g.SpeedGreaterThanEachOtherForEffect(opp) {
		g.ReduceSpeedForEffect(opp, 1, 1)
	}
	return nil
})

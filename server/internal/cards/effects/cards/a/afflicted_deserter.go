package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Afflicted Deserter // Werewolf Ransacker — {3}{R} Creature — Human
// Werewolf 3/2 // Creature — Werewolf 5/4 (#4071, ADR 0132 amendment
// 2026-10-16):
//
//	Front: "At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Whenever this creature transforms into Werewolf Ransacker,
//	        you may destroy target artifact. If that artifact is put into
//	        a graveyard this way, this creature deals 3 damage to that
//	        artifact's controller.
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// The transform trigger watches the turn-over to the back face only
// (CR 701.27e); a card that enters on its back face never transformed.
// The artifact is targeted as the trigger goes on the stack and the
// "you may" is asked as it resolves (CR 603.5). The damage is dealt only
// if the artifact reached a graveyard (an indestructible one, or one a
// replacement sent elsewhere, earns nothing), to the player who
// controlled it as it was destroyed, and it is dealt even if the
// Ransacker has left the battlefield by then.
//
// No simplification.
func init() {
	const oracle = "52adc761-1a0a-43d1-a706-41a92bb7880c"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Afflicted Deserter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Afflicted Deserter"),
		},
	})
	ransack := effects.On(game.EventTransform, effects.ThisTransformedInto(1),
		"Werewolf Ransacker — you may destroy target artifact", werewolfRansackerDestroy)
	ransack.Targets = effects.TargetPermanent("target artifact", effects.Artifact())
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Werewolf Ransacker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(ransack, "Werewolf Ransacker — destroy target artifact?"),
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Werewolf Ransacker"),
		},
	})
}

// werewolfRansackerDestroy destroys the target and, if it landed in a
// graveyard, has the Ransacker deal 3 damage to its controller.
func werewolfRansackerDestroy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	controller := c.Controller
	return g.DestroyPermanentsKeyedForEffect([]uuid.UUID{target}, effects.WithCarry(effects.NewContext(g, item).ItemTail(afflictedDeserterDestroyed), controller))
}

// afflictedDeserterDestroyed deals 3 to the destroyed artifact's controller (#4253 slice 9).
var afflictedDeserterDestroyed = effects.CardTail("afflicted-deserter/damage-controller-if-destroyed", func(ctx *effects.Context, o game.TailOutcome) error {
	g, item, destroyed := ctx.Game, ctx.Item, o.IDs
	controller, source := effects.TailCarry0(o), ctx.Source()

	if len(destroyed) == 0 {
		return nil
	}
	return effects.DealDamage{Source: source, Target: controller, Amount: 3}.Apply(effects.NewContext(g, item))
})

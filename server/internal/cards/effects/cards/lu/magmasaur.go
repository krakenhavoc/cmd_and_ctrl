package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Magmasaur — Creature — Elemental Dinosaur {3}{R}{R}, 0/0 (Tempest, S82
// OLD batch 760, #5628):
//
//	"This creature enters with five +1/+1 counters on it.
//	 At the beginning of your upkeep, you may remove a +1/+1 counter from
//	 this creature. If you don't, sacrifice this creature and it deals
//	 damage equal to the number of +1/+1 counters on it to each creature
//	 without flying and each player."
//
// The upkeep ask is the Wolfbat/Bane yes-no (MayYesNo): yes removes a counter;
// no (or nothing to remove) sacrifices it and it deals damage equal to the
// counters it had, last-known (CR 608.2h), to each creature without flying
// and each player at once (Flamebreak's sweep). Both branches need the Magmasaur
// to still be the permanent that triggered (CR 400.7): one that has left is
// neither asked nor sacrificed, and deals nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da38bfdf-b4c8-4ffe-ad53-b42c317cae5f",
		Name:         "Magmasaur",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 5, "Magmasaur: enters with five +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Magmasaur — remove a +1/+1 counter, or sacrifice it and deal damage equal to its counters to each creature without flying and each player",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
						return nil
					}
					if c, ok := ctx.Game.LookupCardForEffect(ctx.Source()); !ok || c.Counters[game.CounterPlusOne] <= 0 {
						return magmasaurBlowsUp(ctx)
					}
					return effects.MayChoice{
						Question: "Magmasaur — remove a +1/+1 counter from it? (If you don't, sacrifice it and it deals damage equal to its counters to each creature without flying and each player.)",
						YesLabel: "Remove a +1/+1 counter",
						NoLabel:  "Sacrifice it",
						Then:     magmasaurAnswer,
					}.Apply(ctx)
				}))),
		},
	})
}

// magmasaurAnswer is the controller's answer, keyed: yes removes a counter,
// no sacrifices and blows up. An on-disk key: never renamed, never reused.
var magmasaurAnswer = effects.MayYesNo("confirm/s82-old-b760-magmasaur",
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
			return nil
		}
		if c, ok := ctx.Game.LookupCardForEffect(ctx.Source()); !ok || c.Counters[game.CounterPlusOne] <= 0 {
			return magmasaurBlowsUp(ctx)
		}
		if effects.SourceIsNewObject(ctx.Game, ctx.Item) {
			return nil
		}
		return ctx.Game.AddCounterForEffect(ctx.Source(), game.CounterPlusOne, -1)
	},
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
			return nil
		}
		return magmasaurBlowsUp(ctx)
	})

// magmasaurBlowsUp sacrifices the Magmasaur and has it deal damage equal to
// the +1/+1 counters it had to each creature without flying and each player.
func magmasaurBlowsUp(ctx *effects.Context) error {
	n := effects.CountersOnThisAsItResolves(game.CounterPlusOne)(ctx)
	if err := (effects.SacrificePermanent{Target: ctx.Source()}).Apply(ctx); err != nil {
		return err
	}
	return effects.DamageEachMatchingAndEachPlayerThen(ctx, effects.And(effects.Creature(), effects.WithoutKeyword("flying")), n, effects.RecipientThen{})
}

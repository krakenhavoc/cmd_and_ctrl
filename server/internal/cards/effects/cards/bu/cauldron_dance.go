package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cauldron Dance — Instant {4}{B}{R} (Commander 2017, #5533):
//
//	"Cast this spell only during combat.
//	 Return target creature card from your graveyard to the battlefield.
//	 That creature gains haste. Return it to your hand at the beginning of
//	 the next end step.
//	 You may put a creature card from your hand onto the battlefield. That
//	 creature gains haste. Its controller sacrifices it at the beginning of
//	 the next end step."
//
// Two halves, in printed order. The first is Apprentice Necromancer's
// return and haste with Nahiri's return to hand at the next end step, bound
// to the returned object (CR 603.7a): a creature that has left and come back
// is not returned again. The second is Sneak Attack's optional put from hand
// with haste and the end-step sacrifice; the creature enters under your
// control, so you sacrifice it. A target that is gone makes the whole spell
// do nothing (CR 608.2b), the second half included. The timing is any step
// of the combat phase (CR 506.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9784eb4-96f2-409e-a7b5-eb6478644943",
		Name:         "Cauldron Dance",
		Completeness: effects.CompletenessFull,
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			return effects.InCombatStep(g.Turn.Step)
		},
		CastConditionLabel: "Cast this spell only during combat.",
		Targets:            effects.TargetCreatureInYourGraveyard(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			obj, ok, err := effects.S82OldB233ReturnWithHaste(ctx, t.ID)
			if err != nil {
				return err
			}
			if ok {
				if err := (effects.ScheduleDelayedTrigger{
					At:     game.StepEnd,
					Label:  "Cauldron Dance — return the creature to your hand",
					Cards:  []uuid.UUID{obj.ID},
					Body:   effects.ZaraReturnToHandBody,
					Params: game.EffectParams{Object: obj},
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.PutFromHandOntoBattlefield{
				Match:    effects.Creature(),
				Optional: true,
				Label:    "Cauldron Dance — you may put a creature card from your hand onto the battlefield",
				ThenTail: cauldronDanceEntered,
			}.Apply(effects.NewContext(ctx.Game, item))
		},
	})
}

// cauldronDanceEntered is "that creature gains haste. Its controller
// sacrifices it at the beginning of the next end step". An on-disk key.
var cauldronDanceEntered = effects.PutFromHandTail("cauldron-dance/entered", func(ctx *effects.Context, res effects.PutFromHandResult) error {
	if res.Entered == uuid.Nil {
		return nil
	}
	c, ok := effects.BattlefieldCardForEffect(ctx.Game, res.Entered)
	if !ok {
		return nil
	}
	if err := (effects.ScopedEffectFor{Target: res.Entered, Mods: []game.Mod{game.AddKeywordsMod("haste")},
		Duration: game.IndefiniteDuration(), Label: "Cauldron Dance — haste"}).Apply(ctx); err != nil {
		return err
	}
	return effects.ScheduleDelayedTrigger{
		Label:  "Cauldron Dance — sacrifice the creature",
		Body:   effects.StdB66MeekAttackSacrifice,
		Params: game.EffectParams{Object: game.ObjectRef{ID: res.Entered, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})

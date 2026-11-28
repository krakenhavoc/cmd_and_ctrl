package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jinxed Choker — Artifact {3} (Mirrodin, S82 OLD, #5606):
//
//	"At the beginning of your end step, target opponent gains control of
//	 this artifact and puts a charge counter on it.
//	 At the beginning of your upkeep, this artifact deals damage to you
//	 equal to the number of charge counters on it.
//	 {3}: Put a charge counter on this artifact or remove one from it."
//
// Jinxed Idol's handover at the end step, with the counter placed after it
// (both only if the Choker is still the permanent the trigger came from,
// CR 400.7). The upkeep trigger is read as it resolves from the counters
// the Choker has, or had if it left (CR 608.2h), and it is the damage
// source, so its controller's prevention sees an artifact. The activated
// ability is the controller's choice made as it resolves (an option
// pick): add one, or remove one, and removing with none is nothing.
//
// No simplification.
func init() {
	const name = "Jinxed Choker"
	effects.Register(effects.Spec{
		OracleID:     "3241abc1-d739-4ac3-93d1-7fc2898b943b",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtYourEndStep(name+" — target opponent gains control of it and puts a charge counter on it", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				targets := ctx.LegalTargets()
				if len(targets) == 0 || !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				if err := (effects.GainControl{
					Target:     item.SourceCardID,
					Controller: targets[0].ID,
					Duration:   game.IndefiniteDuration(),
					Label:      name + " — target opponent gains control of it",
				}).Apply(ctx); err != nil {
					return err
				}
				return effects.CounterOnThis{Kind: game.CounterCharge, N: 1}.Apply(ctx)
			}), effects.TargetPlayer("target opponent", effects.Opponent())),
			effects.AtYourUpkeep(name+" — it deals damage to you equal to the number of charge counters on it", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				if !ok || info.Counters[game.CounterCharge] <= 0 {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: info.Counters[game.CounterCharge]}.Apply(ctx)
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Put a charge counter on this artifact or remove one from it.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				g.QueueOptionPickForEffect(game.OptionPickPrompt{
					Chooser:  item.Controller,
					Source:   item.SourceCardID,
					Question: name + " — put a charge counter on it or remove one from it?",
					Options: []game.ChoiceOption{
						{Label: "Put a charge counter on it"},
						{Label: "Remove a charge counter from it"},
					},
					ThenKey: jinxedChokerCounter,
					Carry:   []uuid.UUID{item.SourceCardID},
				})
				return nil
			},
		}},
	})
}

// jinxedChokerCounter adds or removes the charge counter chosen, while the
// Choker is still on the battlefield. An on-disk key: never renamed, never
// reused.
var jinxedChokerCounter = game.RegisterOptionPickThen("option-pick/s82-old-b603-jinxed-choker-counter", func(g *game.Game, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) == 0 || !effects.OnBattlefield(g, r.Carry[0]) {
		return nil
	}
	if r.Index == 0 {
		return g.AddCounterForEffect(r.Carry[0], game.CounterCharge, 1)
	}
	if c, ok := g.LookupCardForEffect(r.Carry[0]); !ok || c.Counters[game.CounterCharge] <= 0 {
		return nil
	}
	return g.AddCounterForEffect(r.Carry[0], game.CounterCharge, -1)
})

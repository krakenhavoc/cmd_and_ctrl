package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wolverine, Best There Is — {1}{R}{G} Legendary Creature — Mutant
// Berserker Hero 2/2 (Marvel's Spider-Man, #4460):
//
//	"Unrivaled Lethality — Double all damage Wolverine would deal.
//	 At the beginning of each end step, if Wolverine dealt damage to
//	 another creature this turn, put a +1/+1 counter on him.
//	 {1}{G}: Regenerate Wolverine."
//
// Unrivaled Lethality is Mjölnir's doubling replacement (CR 614.1a) with
// the source narrowed to Wolverine himself, combat or not.
//
// The end-step trigger is an intervening "if" (CR 603.4): it reads the
// turn's damage records (TurnTally.DamagedObjects), which name the
// Wolverine OBJECT and mark whether the permanent he hit was a creature
// as the damage was dealt. It is checked as the end step begins and
// again as the trigger resolves, against the Wolverine on the
// battlefield then; one who left and came back is a new object that has
// dealt no damage (CR 400.7).
//
// The regeneration is Albino Troll's (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6e53a43e-5a23-4057-9bb5-281508da57af",
		Name:         "Wolverine, Best There Is",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventDamage && ev.DamageAmount > 0 &&
					ev.DamageSource != uuid.Nil && ev.DamageSource == src.InstanceID
			},
			Replace:    b39DoubleTheDamage,
			Controller: b39SourceController,
			Label:      "Wolverine, Best There Is: double all damage Wolverine would deal",
		}},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "Wolverine dealt damage to another creature this turn",
				Test: func(ctx *Context) bool { return wolverineDealtDamageToAnotherCreature(ctx.Game, ctx.Source()) },
			}, On(game.EventBeginEndStep,
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return wolverineDealtDamageToAnotherCreature(g, source.InstanceID)
				},
				"Wolverine, Best There Is — put a +1/+1 counter on him",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: Regenerate Wolverine.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{1}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}

// wolverineDealtDamageToAnotherCreature is the intervening "if": the
// Wolverine object on the battlefield now dealt damage this turn to a
// permanent that was a creature and was not him. Off the battlefield he
// is no object that can have.
func wolverineDealtDamageToAnotherCreature(g *game.Game, id uuid.UUID) bool {
	c, ok := battlefieldCardForEffect(g, id)
	if !ok {
		return false
	}
	return g.ObjectDealtDamageToAnotherCreatureThisTurn(game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
}

package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Clone Saga — {3}{U} Enchantment — Saga (Marvel Super Heroes Commander,
// #4455):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Surveil 3.
//	 II — When you next cast a creature spell this turn, copy it, except the
//	 copy isn't legendary.
//	 III — Choose a card name. Whenever a creature with the chosen name deals
//	 combat damage to a player this turn, draw a card."
//
// II is Summon: G.F. Cerberus's "when you next cast" delayed trigger over a
// creature spell (CR 603.7), whose body is Double Major's copy with the
// legendary supertype removed from the copy (CR 707.10a). The copy of a
// creature spell becomes a token as it resolves (CR 707.10, 608.3f). III asks
// for the name as it resolves (Demonic Consultation's prompt), then schedules
// a delayed trigger for the rest of the turn (CR 603.7) that fires once for
// each creature with that name, matched on the creature as it dealt the
// damage; the name rides the trigger as plain data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0f0282c6-aceb-4879-8ec8-482d0501204a",
		Name:         "The Clone Saga",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel("The Clone Saga", 1, "surveil 3"), Do(Surveil{N: 3})),
			ChapterTrigger(2, SagaChapterLabel("The Clone Saga", 2, "when you next cast a creature spell this turn, copy it, except the copy isn't legendary"),
				func(g *game.Game, item *game.StackItem) error {
					return WhenYouNextCast("The Clone Saga — copy that creature spell, except the copy isn't legendary",
						game.CastFilter{Types: []string{"Creature"}}, cloneSagaCopyBody).Apply(NewContext(g, item))
				}),
			ChapterTrigger(3, SagaChapterLabel("The Clone Saga", 3, "choose a card name; whenever a creature with that name deals combat damage to a player this turn, draw a card"),
				func(g *game.Game, item *game.StackItem) error {
					g.QueueCardNameChoiceKeyedForEffect(item.Controller, item.SourceCardID,
						"The Clone Saga — choose a card name", NewContext(g, item).ItemTail(cloneSagaNamed))
					return nil
				}),
		},
	})
}

var (
	// "Copy it, except the copy isn't legendary." On-disk identity.
	cloneSagaCopyBody = game.SimpleDelayedBody("clone-saga/copy-creature-spell-not-legendary", cloneSagaCopy)

	// "Whenever a creature with the chosen name deals combat damage to a
	// player this turn". On-disk identity.
	cloneSagaNamedHitsCondition = game.DelayedCondition("clone-saga/named-creature-combat-damage-to-player",
		func(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
			if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || p.Name == "" || g.PlayerByIDForEffect(ev.Target) == nil {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.Source)
			return ok && c.IsCreature() && game.CardNameMatches(c, p.Name)
		})

	// The rest of chapter III once the name (the tail's Word) is known. An
	// on-disk key.
	cloneSagaNamed = CardTail("then/clone-saga-named", func(ctx *Context, o game.TailOutcome) error {
		return DelayedOnEvent{
			Label:      "The Clone Saga — draw a card (" + o.Tail.Word + " dealt combat damage to a player)",
			On:         []game.EventKind{game.EventDealDamage},
			Condition:  cloneSagaNamedHitsCondition,
			CondParams: game.EffectParams{Name: o.Tail.Word},
			Body:       drawOneBody,
			Repeats:    true,
		}.Apply(ctx)
	})
)

func cloneSagaCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cast := ctx.PayloadCards()
	if len(cast) == 0 {
		return nil
	}
	return CopySpell{
		StackID:          cast[0],
		Controller:       item.Controller,
		ChooseNewTargets: true,
		FromLastKnown:    true,
		Except: func(v *game.PrintedValues) {
			v.RemoveSupertype("Legendary")
		},
	}.Apply(ctx)
}

package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fiery Encore — Sorcery {4}{R} (Commander 2021, S82 OLD b058, #5504):
//
//	"Discard a card, then draw a card. When you discard a nonland card this
//	 way, Fiery Encore deals damage equal to that card's mana value to target
//	 creature or planeswalker.
//	 Storm (When you cast this spell, copy it for each spell cast before it
//	 this turn.)"
//
// A rummage: the discard is mandatory (an empty hand discards nothing), then
// the draw lands. A nonland discard creates the reflexive trigger (CR 603.12)
// once the draw is done, so it goes on the stack after the spell finishes
// resolving: its target is chosen then, and the damage is Fiery Encore's, as
// it last existed, equal to the discarded card's mana value read where it
// lies now (Glacial Dragonhunt's shape with the amount in Params). Each storm
// copy rummages and triggers on its own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "921e7f6b-bf49-46f1-9fdd-b0e5fde5a5a1",
		Name:         "Fiery Encore",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Storm()},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player:   item.Controller,
				Source:   item.SourceCardID,
				N:        1,
				Question: "Fiery Encore — discard a card",
			}, fieryEncoreDiscarded, nil)
		},
	})
}

var (
	// The discard's answer, keyed: draw a card, then a nonland discard creates
	// the reflexive trigger. The key is an on-disk identity.
	fieryEncoreDiscarded = game.RegisterDiscardThen("choose-cards/fiery-encore-discarded", func(g *game.Game, r game.DiscardsDone) error {
		mv, nonland := 0, false
		for _, id := range r.Discarded.Cards() {
			if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
				nonland, mv = true, c.ManaValue()
			}
		}
		if err := g.DrawNForEffect(r.Player, 1); err != nil {
			return err
		}
		if !nonland {
			return nil
		}
		parent := &game.StackItem{Controller: r.Player, SourceCardID: r.Source, Label: "Fiery Encore"}
		t := effects.WhenYouDo("Fiery Encore — damage equal to the discarded card's mana value to target creature or planeswalker", fieryEncoreDamageBody)
		t.Params = game.EffectParams{Amount: mv}
		return t.Apply(effects.NewContext(g, parent))
	})

	// "Fiery Encore deals damage equal to that card's mana value to target
	// creature or planeswalker." Breeches' damage body, the value in
	// Params.Amount.
	fieryEncoreDamageBody = game.ReflexiveBody("fiery-encore/damage", effects.BreechesBlastEffect,
		effects.ConstTargets(func() *game.TargetSpec {
			return effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))
		}))
)

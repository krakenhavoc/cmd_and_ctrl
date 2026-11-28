package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Broken Visage — Instant {4}{B} (Homelands, S82 OLD b827, #5635):
//
//	"Destroy target nonartifact attacking creature. It can't be regenerated.
//	 Create a black Spirit creature token. Its power is equal to that
//	 creature's power and its toughness is equal to that creature's toughness.
//	 Sacrifice the token at the beginning of the next end step."
//
// Befoul's destruction with regeneration refused (CR 701.19c). The creature's
// power and toughness are read through the layers as the spell resolves, before
// it is destroyed (CR 608.2h: last-known information), and the Spirit is made
// from the destruction's keyed tail, so it comes after the creature has gone
// whether or not it did (an indestructible creature still gives the Spirit its
// size). The token is a 0/0 row set to those numbers as it is created, as every
// X/X token is, and a delayed trigger (CR 603.7) sacrifices it at the next end
// step. The target is checked at announce and again on resolution (CR 608.2b):
// one that is no longer a nonartifact attacking creature takes the whole spell
// with it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e1d750e-44a7-47b1-8e0a-5f8b15d48b41",
		Name:         "Broken Visage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1},
		Targets:      effects.TargetCreature("target nonartifact attacking creature", effects.Not(effects.Artifact()), effects.AttackingCreature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			c, found := ctx.Game.LookupCardForEffect(t.ID)
			if !found {
				return nil
			}
			tail := ctx.ItemTail(brokenVisageSpirit)
			tail.Nums = []int{c.Power, c.Toughness}
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{t.ID}, tail, game.DestroyOptions{CantBeRegenerated: true})
		},
	})
}

var (
	// brokenVisageSpirit is "Create a black Spirit creature token. Its power is
	// equal to that creature's power ...": the numbers ride the tail. An on-disk
	// key: never renamed, never reused.
	brokenVisageSpirit = effects.CardTail("then/old-b827-broken-visage-create-spirit", func(ctx *effects.Context, o game.TailOutcome) error {
		if len(o.Tail.Nums) < 2 {
			return nil
		}
		tok := ctx.TextToken(effects.TokenCard("0/0 black Spirit"))
		tok.Power, tok.Toughness = o.Tail.Nums[0], o.Tail.Nums[1]
		return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
			Controller: ctx.Controller(), Source: ctx.Source(),
			Groups: []game.TokenGroup{{Template: tok, Count: 1}},
		}, ctx.ItemTail(brokenVisageSacrifice))
	})
	// brokenVisageSacrifice schedules the Spirit's sacrifice at the next end
	// step. An on-disk key: never renamed, never reused.
	brokenVisageSacrifice = effects.CardTail("then/old-b827-broken-visage-sacrifice-spirit", func(ctx *effects.Context, o game.TailOutcome) error {
		if len(o.IDs) == 0 {
			return nil
		}
		return effects.ScheduleDelayedTrigger{
			Label: "Broken Visage — sacrifice the Spirit token",
			Cards: o.IDs,
			Body:  effects.SacrificeListedCardsBody,
		}.Apply(ctx)
	})
)

package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Horn of Plenty — Artifact {6} (Mercadian Masques, S82 OLD, #5619):
//
//	"Whenever a player casts a spell, they may pay {1}. If the player does,
//	 they draw a card at the beginning of the next end step."
//
// Whoever cast the spell is asked, not Horn's controller (Rhystic Study's
// asker, for any caster). Paying schedules a delayed trigger (CR 603.7) for the
// next end step at the table; Horn's controller controls it (CR 603.7d) but the
// caster draws, so the caster rides its params. A caster who has left the game
// by then draws nothing. A payment the caster can't fund is a decline.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "72d3b499-dbff-4631-b6e3-0a292b39d527",
		Name:         "Horn of Plenty",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, hornOfPlentyAnyCast,
				"Horn of Plenty — that player may pay {1}; if they do, they draw a card at the beginning of the next end step",
				hornOfPlentyAsk),
		},
	})
}

func hornOfPlentyAnyCast(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor != uuid.Nil
}

func hornOfPlentyAsk(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	caster := item.Trigger.Event.Actor
	return effects.MayPay{
		Chooser:  caster,
		Cost:     "{1}",
		Question: "Horn of Plenty — pay {1} to draw a card at the beginning of the next end step?",
		Then:     hornOfPlentyPaid,
	}.Apply(effects.NewContext(g, item))
}

// hornOfPlentyPaid schedules the caster's draw. An on-disk key: never renamed,
// never reused.
var hornOfPlentyPaid = effects.PayUnlessBranches("pay-unless/horn-of-plenty-draw",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		return effects.ScheduleDelayedTrigger{
			Label:  "Horn of Plenty — draw a card",
			Body:   hornOfPlentyDrawBody,
			Params: game.EffectParams{Player: r.Chooser},
		}.Apply(ctx)
	}, nil)

// hornOfPlentyDrawBody is the delayed trigger: the player in its params draws a
// card. An on-disk key: never renamed, never reused.
var hornOfPlentyDrawBody = game.DelayedBody("horn-of-plenty/draw",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		if p.Player == uuid.Nil || g.PlayerByIDForEffect(p.Player) == nil {
			return nil
		}
		return effects.DrawCards{Player: p.Player, N: 1}.Apply(effects.NewContext(g, item))
	})

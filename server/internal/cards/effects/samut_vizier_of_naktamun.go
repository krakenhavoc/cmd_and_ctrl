package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Samut, Vizier of Naktamun — Legendary Creature — Human Warrior Cleric
// {1}{R}{G}, 2/3 (March of the Machine: The Aftermath):
//
//	"First strike, vigilance, haste
//	 Whenever a creature you control deals combat damage to a player, if
//	 that creature entered this turn, draw a card."
//
// Bident of Thassa's combat-damage trigger (one event per creature, so
// three new creatures connecting is three draws) with an intervening "if"
// (CR 603.4) on the damaging creature: it must have entered the battlefield
// this turn, read from the per-object entered-this-turn tally (so a creature
// that entered on an opponent's turn, or one that left and came back, is
// judged as the object that dealt the damage). Checked as the ability
// triggers and again as it resolves, off the event's source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "178b3fdf-dbee-4245-a811-b21ff8a24f66",
		Name:            "Samut, Vizier of Naktamun",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "vigilance", "haste"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(samutDamagerEnteredThisTurn, On(game.EventDealDamage,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return combatDamageToPlayerBy(ev, source.Controller, g) && b06EnteredThisTurn(g, ev.Source)
				},
				"Samut, Vizier of Naktamun — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}

// samutDamagerEnteredThisTurn is the intervening "if that creature entered
// this turn", read for the creature the trigger's event names.
var samutDamagerEnteredThisTurn = CondBoard{
	Name: "that creature entered this turn",
	Test: func(ctx *Context) bool { return b06EnteredThisTurn(ctx.Game, ctx.Trigger().Event.Source) },
}

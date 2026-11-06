package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Krang, the All-Powerful — Legendary Artifact Creature — Utrom Robot
// {4}{U}, 3/3 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"If a player drawing a card causes a triggered ability of a
//	 permanent you control to trigger, that ability triggers an
//	 additional time.
//	 Whenever a player draws their second card each turn, put a +1/+1
//	 counter on Krang."
//
// The first ability is a CR 603.2d doubler (tmcB1bDoublesDrawing): an
// EventDrawCard, by any player, that triggers an ability of a permanent
// Krang's controller controls adds one more instance, with its own
// choices and targets (the ruling). A spell's, an emblem's, a card's in
// another zone and a delayed trigger are not a permanent's ability.
//
// The second watches every player's draws and reads the drawer's own
// turn tally after the draw (King T'Challa's reading), so it fires once
// per player per turn, on that player's second card. That draw is a
// player drawing a card, so Krang's own trigger is doubled by Krang: two
// counters.
//
// No simplification.
func init() {
	const name = "Krang, the All-Powerful"
	Register(Spec{
		OracleID:        "466d5226-f4c7-4d69-9f56-4f893010127f",
		Name:            name,
		Completeness:    CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{tmcB1bDoublesDrawing(name)},
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor != uuid.Nil && g.TurnTallyFor(ev.Actor).CardsDrawn == 2
			}, "Krang, the All-Powerful — put a +1/+1 counter on Krang (a player drew their second card)",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sengir, the Dark Baron — {4}{B}{B} Legendary Creature — Vampire Noble
// 4/4 (Commander Legends, S82 OLD b077, #6164):
//
//	"Flying
//	 Whenever another creature dies, put two +1/+1 counters on Sengir.
//	 Whenever another player loses the game, you gain life equal to that
//	 player's life total as the turn began.
//	 Partner (You can have two commanders if both have partner.)"
//
// The life gain reads #6164's record of each player's life as the turn
// began, which the engine stamps on the player as every turn begins and
// copies onto the player's elimination event (Event.Amount): a player at
// 12 as the turn began who loses at −3 gives Sengir's controller 12. A
// loss to zero or less life as the turn began gains nothing. Partner is a
// deck-construction rule (CR 702.124a) and needs nothing on the card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f49f84f9-616e-495b-8d87-4de7941f230e",
		Name:            "Sengir, the Dark Baron",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B15AnotherCreatureDied(ev, source, g)
				},
				Key:    "Sengir, the Dark Baron — put two +1/+1 counters on Sengir",
				Effect: effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 2}),
			},
			effects.On(game.EventPlayerEliminated, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != source.Controller
			}, "Sengir, the Dark Baron — gain life equal to that player's life total as the turn began",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					n := ctx.Trigger().Event.Amount
					if n <= 0 {
						return nil
					}
					return effects.GainLife{Amount: n}.Apply(ctx)
				}))),
		},
	})
}

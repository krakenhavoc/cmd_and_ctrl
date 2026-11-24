package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Relic Seeker — Creature — Human Soldier {1}{W}, 2/2 (Commander Legends):
//
//	"Renown 1 (When this creature deals combat damage to a player, if it isn't
//	 renowned, put a +1/+1 counter on it and it becomes renowned.)
//	 When this creature becomes renowned, you may search your library for an
//	 Equipment card, reveal it, put it into your hand, then shuffle."
//
// Renown is the engine's keyword trigger; the search watches
// EventBecameRenowned for this creature alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1f101363-c7d1-405a-b3a9-e93b46928dbf",
		Name:            "Relic Seeker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"renown 1"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventBecameRenowned, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && ev.CardID == source.InstanceID
			}, "Relic Seeker — you may search your library for an Equipment card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtype("Equipment") },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Relic Seeker — you may search your library for an Equipment card",
				}.Apply(effects.NewContext(g, item))
			}), game.Purpose{Tutors: 1}),
		},
	})
}

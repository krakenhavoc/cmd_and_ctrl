package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// River Kelpie — Creature — Beast {3}{U}{U}, 3/3:
//
//	"Whenever this creature or another permanent enters from a
//	 graveyard, draw a card.
//	 Whenever a player casts a spell from a graveyard, draw a card.
//	 Persist"
//
// Both triggers draw for the Kelpie's controller, whoever returned or
// cast the card. The first reads EventETB.EnteredFrom (ADR 0113 amendment
// 2026-10-08); the second reads EventCast.OldZone, the zone the spell
// was cast from. "Another permanent" is any type, so a land returned
// from a graveyard counts; a token does not, it comes from no zone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ef5c93ae-91e0-4edb-b8a3-1171a3694bb8",
		Name:            "River Kelpie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.EnteredFromAGraveyard,
				"River Kelpie — draw a card",
				effects.Do(effects.DrawCards{N: 1})),
			effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventCast && ev.OldZone == game.ZoneGraveyard
			}, "River Kelpie — draw a card (spell cast from a graveyard)",
				effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

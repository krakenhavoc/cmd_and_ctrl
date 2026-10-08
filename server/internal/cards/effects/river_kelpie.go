package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "ef5c93ae-91e0-4edb-b8a3-1171a3694bb8",
		Name:            "River Kelpie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, EnteredFromAGraveyard,
				"River Kelpie — draw a card",
				Do(DrawCards{N: 1})),
			On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventCast && ev.OldZone == game.ZoneGraveyard
			}, "River Kelpie — draw a card (spell cast from a graveyard)",
				Do(DrawCards{N: 1})),
		},
	})
}

package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gilded Light — Instant {1}{W} (Scourge, #5842):
//
//	"You gain shroud until end of turn. (You can't be the target of
//	 spells or abilities.)
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Player shroud is the player-keyword grant (player_keywords.go), as
// Lazotep Plating grants hexproof, until end of turn (CR 514.2). The
// targeting gate reads it per CR 702.18a: no spell or ability can target
// you, your own included. Cycling is CR 702.29a.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7e7b2aa-9fd2-41dd-b11e-3a233530dc83",
		Name:         "Gilded Light",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GainPlayerKeyword{
				Player:   ctx.Controller(),
				Keyword:  effects.KeywordShroud,
				Label:    "Gilded Light — you gain shroud",
				Duration: effects.DurationUntilEndOfTurn(ctx),
			}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}

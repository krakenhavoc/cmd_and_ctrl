package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ignite the Future — Sorcery {3}{R}:
//
//	"Exile the top three cards of your library. Until the end of your next
//	 turn, you may play those cards. If this spell was cast from a graveyard,
//	 you may play cards this way without paying their mana costs.
//	 Flashback {7}{R}"
//
// The window is "until the end of your next turn" (CR 514.2: it ends at that
// turn's cleanup). A spell cast with flashback remembers the zone it came
// from (CR 601.2a), and then the permission is free; a land is played, not
// cast, so it is played either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "915aa871-4c5e-4799-9a07-d7b855df4e7a",
		Name:             "Ignite the Future",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerValue},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{7}{R}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			window := ctx.Game.UntilEndOfYourNextTurnDuration(item.Controller)
			return effects.ExileTopWithPermission{
				From:    item.Controller,
				GrantTo: item.Controller,
				N:       3,
				Window:  &window,
				Free:    item.CastFromZone == game.ZoneGraveyard,
			}.Apply(ctx)
		},
	})
}

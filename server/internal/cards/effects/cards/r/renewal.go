package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Renewal — Sorcery {2}{G} (Homelands, S82 OLD b826, #5635):
//
//	"As an additional cost to cast this spell, sacrifice a land.
//	 Search your library for a basic land card, put that card onto the
//	 battlefield, then shuffle.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// The land is sacrificed while casting (CR 601.2h), so a spell that is
// countered still cost the land and a player with no land can't cast it. The
// search puts the basic land onto the battlefield untapped, then shuffles; the
// draw is the delayed trigger for the next upkeep at the table (CR 603.7a,
// next_upkeep_draw.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "8df46910-ee13-416e-8282-dfe85a90e6d5",
		Name:           "Renewal",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Lands: 1, LandsUntapped: 1, Draws: 1},
		AdditionalCost: effects.SacrificeCost("a land", effects.Land()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: effects.IsBasicLand,
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Renewal — a basic land card onto the battlefield",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Renewal — draw a card").Apply(ctx)
		},
	})
}

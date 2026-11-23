package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Path to the Festival — Sorcery {2}{G} (Midnight Hunt, #5497):
//
//	"Search your library for a basic land card, put that card onto the
//	 battlefield tapped, then shuffle. Then if there are three or more
//	 basic land types among lands you control, scry 1.
//	 Flashback {4}{G} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Rampant Growth's search with the "then if" as the search's tail, so the
// count is taken after the land has entered and the library is shuffled
// (and after a search prompt, which makes the wait a restore point). The
// count is DomainCount, which reads the layered lands. The scry is
// unconditional on what the search found: a failed search still scries on
// three types.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "e54d4ecf-ebcc-4374-98e7-c682f83c82a9",
		Name:             "Path to the Festival",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Lands: 1, Scry: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{4}{G}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     effects.IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Path to the Festival — a basic land",
				ThenTail:      ctx.Tail(pathToTheFestivalScry),
			}.Apply(ctx)
		},
	})
}

var pathToTheFestivalScry = effects.CardTail("path-to-the-festival/scry-on-three-types", func(ctx *effects.Context, o game.TailOutcome) error {
	if effects.DomainCount(ctx.Game, o.Tail.Player) < 3 {
		return nil
	}
	return effects.Scry{Player: o.Tail.Player, N: 1}.Apply(ctx)
})

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cultivate — "Search your library for up to two basic land cards,
// reveal those cards, put one onto the battlefield tapped and the
// other into your hand, then shuffle."
//
// Two sequential searches rather than one, because the two halves go
// to different zones and the primitive has one destination. The
// first defers its shuffle so the second still sees the library;
// the second shuffles, which is the single shuffle the rules ask
// for.
//
// S22: the searcher picks both cards. The second search is CHAINED
// off the first via Then — it cannot run on the line below, because
// the first search now returns while its prompt is still open, and
// a second prompt opened at that moment would offer a card the
// player is in the middle of taking.
//
// "Up to two" degrades the way it reads: a library with none no-ops
// both halves. A library with exactly ONE basic land is the case the
// pair of searches cannot express (the first would take it for the
// battlefield with no say): the one card found goes to the battlefield
// tapped OR to the hand, the searcher's choice, so that card is asked
// about (a confirm prompt) and fetched to the answer's zone, with the
// one shuffle.
//
// The land enters the battlefield TAPPED because Cultivate says so
// (SearchLibrary.TappedOnEntry). Since #263 the fetched land's own
// enters-tapped clause runs as well, so a Cultivated checkland is
// tapped for the printed reason on top of this one.
func init() {
	Register(Spec{
		OracleID:     "8b755881-a72d-4e21-a369-d2924eb4585a",
		Name:         "Cultivate",
		Purpose:      game.Purpose{Lands: 1, Tutors: 1},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			source := ctx.Source()
			if cultivateOnlyOneBasic(ctx, controller) {
				return cultivateLoneBasic(ctx, controller, source, "Cultivate")
			}
			return SearchLibrary{
				Player:        controller,
				Source:        source,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				Shuffle:       false,
				TappedOnEntry: true,
				Reason:        "Cultivate — basic land onto the battlefield tapped",
				Then: func(g *game.Game, _ []uuid.UUID) error {
					return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
						Player:  controller,
						Source:  source,
						Pred:    IsBasicLand,
						Dest:    game.ZoneHand,
						Limit:   1,
						Reveal:  true,
						Shuffle: true,
						Reason:  "Cultivate — basic land into your hand",
					})
				},
			}.Apply(ctx)
		},
	})
}

// cultivateOnlyOneBasic reports whether the player's library holds
// exactly one basic land card: the case where "put one onto the
// battlefield tapped and the other into your hand" has a single card
// to place and the searcher chooses where.
func cultivateOnlyOneBasic(ctx *Context, player uuid.UUID) bool {
	p := ctx.PlayerByID(player)
	if p == nil || p.Library == nil {
		return false
	}
	n := 0
	for _, c := range p.Library.Cards {
		if IsBasicLand(c) {
			n++
		}
	}
	return n == 1
}

// cultivateLoneBasic asks where the single basic goes, then fetches it
// there and shuffles. Shared with Kodama's Reach, which prints the
// same clause.
func cultivateLoneBasic(ctx *Context, controller, source uuid.UUID, name string) error {
	fetch := func(dest game.ZoneKind, reason string) func(g *game.Game) error {
		return func(g *game.Game) error {
			return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player:        controller,
				Source:        source,
				Pred:          IsBasicLand,
				Dest:          dest,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: dest == game.ZoneBattlefield,
				Reason:        reason,
			})
		}
	}
	ctx.Game.QueueConfirmForEffect(game.ConfirmPrompt{
		Chooser:      controller,
		Source:       source,
		Question:     name + " — put the basic land onto the battlefield tapped or into your hand?",
		AcceptLabel:  "Battlefield, tapped",
		DeclineLabel: "Into your hand",
		OnAccept:     fetch(game.ZoneBattlefield, name+" — basic land onto the battlefield tapped"),
		OnDecline:    fetch(game.ZoneHand, name+" — basic land into your hand"),
	})
	return nil
}

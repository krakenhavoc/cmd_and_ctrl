package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ormos, Archive Keeper — Legendary Creature — Sphinx {4}{U}{U}, 5/5
// (Jumpstart Azorius deck request, #4894; seam #4896):
//
//	"Flying
//	 If you would draw a card while your library has no cards in it,
//	 instead put five +1/+1 counters on Ormos.
//	 {1}{U}{U}, Discard three cards with different names: Draw five
//	 cards."
//
// The draw replacement is Laboratory Maniac's window (CR 614.11: it
// applies even though no card could be drawn) with a different
// replacement: the draw is cancelled and five +1/+1 counters are put on
// Ormos instead. CR 121.2 makes "draw five" five draws, so each draw that
// finds the library empty is replaced on its own: five counters for each
// of them, and the draws before them are ordinary. The counters ride the
// draw-instead body (DrawInstead) rather than a Replace, because a counter
// placement can itself pause on a CR 616 ordering prompt (Doubling Season
// beside Hardened Scales); the rest of the draw instruction waits behind
// it as a parked DrawDone. A replaced draw is not a draw, so nothing is
// owed to CR 704.5b.
//
// The cost is a three-card discard with the "different names" set rule
// (DiscardNDifferentNames, #4896): the activation refuses two cards
// sharing a name (CR 201.2b; both halves of a split card name it,
// CR 709.4), the enumerator pays three distinct names or offers nothing,
// and the client's picker greys a repeat.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "06d60b71-a9f9-4a7f-ad69-15d760b6e5c3",
		Name:            "Ormos, Archive Keeper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{ormosCountersInsteadOfDrawing()},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}{U}, Discard three cards with different names: Draw five cards.",
			Cost:    Plus(ManaCost("{1}{U}{U}"), DiscardNDifferentNames(3, "three cards with different names")),
			Purpose: game.Purpose{Draws: 5, Answers: game.AnswerValue},
			Effect:  Do(DrawCards{N: 5}),
		}},
	})
}

// ormosCountersInsteadOfDrawing is "If you would draw a card while your
// library has no cards in it, instead put five +1/+1 counters on Ormos".
// The library is read at each draw, as Laboratory Maniac's is.
func ormosCountersInsteadOfDrawing() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDrawCard},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventDraw || ev.DrawCount <= 0 || src == nil || ev.DrawPlayer != src.Controller {
				return false
			}
			p := g.PlayerByIDForEffect(ev.DrawPlayer)
			return p != nil && !p.Eliminated && p.Library.Size() == 0
		},
		DrawInstead: game.RegisterDrawInstead("ormos-archive-keeper", ormosPutCounters),
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Ormos, Archive Keeper: put five +1/+1 counters on Ormos instead of drawing from an empty library",
	}
}

// ormosPutCounters is the replaced draw: five +1/+1 counters on Ormos,
// placed by the drawer (the effect's controller), then the rest
// of the draw instruction (the parked DrawDone). Ormos gone by now puts
// nothing and the instruction goes on.
func ormosPutCounters(g *game.Game, drawer, source uuid.UUID, done game.DrawDone) error {
	if z := g.FindCardZoneForEffect(source); z == nil || z.Kind != game.ZoneBattlefield {
		return done.Run(g)
	}
	then := game.KeyedTail{Ref: ormosThenDraw, Player: drawer, Carry: []uuid.UUID{done.Park(g)}}
	return g.AddCounterByKeyedForEffect(drawer, source, game.CounterPlusOne, 5, then)
}

// ormosThenDraw pays the rest of the draw once the counters have settled.
// An on-disk key.
var ormosThenDraw = game.RegisterTail("ormos-archive-keeper/then-draw", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return game.DrawDoneByID(o.Tail.Carry[0]).Run(g)
})

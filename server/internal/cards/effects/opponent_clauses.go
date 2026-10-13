package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opponent_clauses.go — "for each opponent, up to one target … that
// player controls" on a SPELL or a reflexive trigger (Windgrace's
// Judgment, Lich's Relic), where Molten Primordial's TargetsFrom is
// not available and the clause list is static card data.
//
// The list is one clause per possible opponent, each bound to the
// opponent in a fixed place in turn order after the chooser: the 1st,
// 2nd or 3rd seat after theirs. A clause for a seat that does not
// exist at this table (a two-player game has only a 1st) offers
// nothing, and every clause is "up to one", so the spell is always
// castable and takes what it can. Binding the pick to its opponent is
// the point: the CR 608.2b re-check asks "does THAT player still
// control it?", so a permanent that changed hands in response is an
// illegal target even when it went to another opponent, which the
// EachDifferentController set rule could not say.

// controlledByNthOpponent passes for a permanent controlled by the
// n-th seat (1-based) after the chooser's, in turn order. Eliminated
// seats keep their place, so the numbering never shifts mid-game.
func controlledByNthOpponent(n int) CardPredicate {
	return func(g *game.Game, chooser uuid.UUID, c game.Card) bool {
		seats := len(g.Seats)
		if n < 1 || n >= seats {
			return false
		}
		at := -1
		for i, p := range g.Seats {
			if p != nil && p.ID == chooser {
				at = i
				break
			}
		}
		if at < 0 {
			return false
		}
		p := g.Seats[(at+n)%seats]
		return p != nil && c.Controller == p.ID
	}
}

var nthOpponentWords = [...]string{"", "first", "second", "third"}

// perOpponentClauses is `count` clauses (at most MaxPlayers-1), the
// n-th matching `what` among permanents the n-th opponent controls.
// upTo makes each "up to one"; otherwise each is exactly one.
func perOpponentClauses(count int, upTo bool, noun string, what ...CardPredicate) *game.TargetSpec {
	if count < 1 {
		count = 1
	}
	if count > game.MaxPlayers-1 {
		count = game.MaxPlayers - 1
	}
	var out []*game.TargetSpec
	for n := 1; n <= count; n++ {
		spec := TargetPermanent("target "+noun+" controlled by the "+nthOpponentWords[n]+" opponent after you in turn order",
			append([]CardPredicate{controlledByNthOpponent(n)}, what...)...)
		if upTo {
			spec = spec.WithCount(0, 1)
		}
		out = append(out, spec)
	}
	return Clauses(out[0], out[1:]...)
}

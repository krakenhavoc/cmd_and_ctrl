package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b110_helpers.go — shared pieces of S82 GAP batch 110.

// gap110ControlledByNthPlayer passes for a permanent controlled by the
// seat n places after the chooser's in turn order; n == 0 is the chooser
// themselves. Eliminated seats keep their place (controlledByNthOpponent's
// rule), so the numbering never shifts mid-game.
func gap110ControlledByNthPlayer(n int) CardPredicate {
	if n > 0 {
		return controlledByNthOpponent(n)
	}
	return func(_ *game.Game, chooser uuid.UUID, c game.Card) bool {
		return c.Controller == chooser
	}
}

// gap110PerPlayerClauses is "for each player, up to one target <noun> that
// player controls": one "up to one" clause per seat at the table, the first
// for the chooser and the rest for the opponents in turn order. A clause for
// a seat that does not exist offers nothing, so the ability always resolves
// and takes what it can.
func gap110PerPlayerClauses(noun string, what ...CardPredicate) *game.TargetSpec {
	words := [...]string{"your", "the first opponent's", "the second opponent's", "the third opponent's"}
	var out []*game.TargetSpec
	for n := 0; n < game.MaxPlayers; n++ {
		spec := TargetPermanent("target "+noun+" controlled by "+words[n]+" seat",
			append([]CardPredicate{gap110ControlledByNthPlayer(n)}, what...)...).WithCount(0, 1)
		out = append(out, spec)
	}
	return Clauses(out[0], out[1:]...)
}

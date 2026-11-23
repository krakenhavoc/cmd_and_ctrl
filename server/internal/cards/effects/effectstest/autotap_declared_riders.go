package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// autotap_declared_riders_test.go — #4189, the follow-up to #4185.
//
// A mana ability's hand-written Rider is a closure the auto-tapper cannot
// read, so a source that carries one is never planned and the move list
// does not count its mana. The riders that say what they do in data are:
// "This permanent deals N damage to you" (PainToYou), "You gain N life"
// (GainLife, #4185) and, here, the two that act on the opponents:
//
//   - OpponentsGainLife: Grove of the Burnwillows' "Each opponent gains
//     1 life";
//   - PingOpponentsTrigger: Zhur-Taa Druid's "deals 1 damage to each
//     opponent", which is a trigger on the stack, not a rider (#4928),
//     and is declared to the planner all the same.
//
// A source that gives the table something is planned, but only after every
// source that does not (tapSource.Table). The sources whose rider spends a
// resource stay manual; the table at the end of the file pins that.

// opponentLife is the life of every seat but p, in seat order.
func OpponentLife(g *game.Game, p *game.Player) []int {
	var out []int
	for _, o := range g.Seats {
		if o.ID != p.ID {
			out = append(out, o.Life)
		}
	}
	return out
}

func SameLife(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

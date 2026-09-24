package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// plot.go — #1342: the plot keyword (CR 702.170a).
//
// "Plot {2}{G}" is one cost, and it is the only thing a card file
// should spell out. Everything else is the keyword's, and the engine
// carries it: the special action's window (your main phase, stack
// empty), the face-up exile, and the plotted state — a free cast from
// exile on a later turn, in the owner's main phase with the stack
// empty, which no flash grant widens (game.TimingPlot, #1318).
//
// "It becomes plotted" on its own — Aven Interrupter's "exile target
// spell. It becomes plotted." — is the PlotExiled primitive, and the
// keyword's performer ends in the same engine call, so a card plotted
// from hand and a card plotted by an effect are indistinguishable.

// Plot is "Plot <cost>" — CR 702.170a. `cost` is the printed plot
// cost, the price of taking the special action; the later cast is
// free (CR 702.170d) and is priced by the grant rather than by an
// offer the caster claims, which is why CastCost stays empty.
//
// Give it to the card the way its oracle text reads:
//
//	SpecialActions: []game.SpecialAction{Plot("{2}{G}")},
func Plot(cost string) game.SpecialAction {
	return game.SpecialAction{
		Kind:  game.SpecialActionPlot,
		Cost:  cost,
		Label: "Plot " + cost,
	}
}

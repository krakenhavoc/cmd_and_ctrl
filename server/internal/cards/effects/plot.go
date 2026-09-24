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

// PlotFromTopOfLibrary is Fblthp, Lost on the Range's two plot
// sentences, which always come together (#1391):
//
//	"The top card of your library has plot. The plot cost is equal to
//	 its mana cost.
//	 You may plot nonland cards from the top of your library."
//
// The controller may plot the nonland card on top of their own library
// for its mana cost, or for its own printed plot cost if it has one.
// Either way it is the ordinary plot special action: sorcery timing,
// face-up exile, and a free cast on a later turn.
//
//	SpecialActionGrants: []game.SpecialActionGrant{PlotFromTopOfLibrary()},
func PlotFromTopOfLibrary() game.SpecialActionGrant {
	return game.SpecialActionGrant{
		Kind:           game.SpecialActionPlot,
		Zone:           game.ZoneLibrary,
		Nonland:        true,
		CostIsManaCost: true,
	}
}

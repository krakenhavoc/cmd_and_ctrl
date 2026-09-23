package game

import "github.com/google/uuid"

// plot.go — "it becomes plotted" (CR 702.170c/d, #1318).
//
//	702.170c …some spells and abilities cause a card in exile to
//	         become plotted.
//	702.170d A plotted card's owner may cast it from exile without
//	         paying its mana cost during their main phase while the
//	         stack is empty during any turn after the turn in which it
//	         became plotted.
//
// Only the EFFECT half exists here: a card already in exile becomes
// plotted (Aven Interrupter's "exile target spell. It becomes
// plotted."). The KEYWORD half — the plot special action that exiles a
// card with plot from its owner's hand (CR 702.170a) — is not built;
// it is #1342.
//
// A plotted card is nothing but a cast permission, which is why this
// is cheap: ADR 0066's CastPermission already carries every clause.
//
//	"its owner may cast it"          Player = owner, CastOnly
//	"from exile"                     Zone = exile, ScopeCards on the
//	                                 instance (so it ends when the card
//	                                 leaves exile, CR 400.7)
//	"without paying its mana cost"   Cost "{0}" — the cascade idiom; an
//	                                 {X} locks at 0 (CR 107.3b)
//	"during their main phase while   Timing = TimingPlot, which no flash
//	 the stack is empty"             grant widens (cast_timing.go)
//	"any turn after the turn in      NotBeforeTurn — see below
//	 which it became plotted"
//
// # The floor
//
// Turn.Number is a ROUND counter, not a turn counter, so "a later
// turn" cannot be written as Number+1 without taking away the owner's
// own turn later in this same round. It does not have to be: the plot
// window only ever opens on the OWNER's turn, so
//
//   - plotted on somebody else's turn, every owner turn from here on is
//     a later turn, and the floor is the current round;
//   - plotted on the owner's own turn, the next owner turn is in the
//     next round, and the floor is the round after.
//
// Declared: an extra turn the owner takes straight after their own, in
// the same round, waits for the next round. Weaker, never stronger.

// PlotExiledCardForEffect makes a card in exile plotted: its owner may
// cast it without paying its mana cost, in their main phase with the
// stack empty, on any later turn, for as long as it stays in exile.
//
// A card that is not in exile is left alone and is not an error — the
// instruction reads "it becomes plotted" off a move that may have gone
// elsewhere (a commander taking CR 903.9's offer), and the caller that
// cares asks through ExileSpellThenForEffect's `exiled`.
//
// Caller must hold g.mu.
func (g *Game) PlotExiledCardForEffect(cardID uuid.UUID, source uuid.UUID) {
	exiled := exiledCardByIDLocked(g, cardID)
	if exiled == nil {
		return
	}
	owner := exiled.Owner
	notBefore := g.Turn.Number
	if g.activeSeatIDLocked() == owner {
		notBefore++
	}
	g.GrantCastPermissionToCardsForEffect(CastPermission{
		Player:        owner,
		Zone:          ZoneExile,
		Cost:          "{0}",
		Timing:        TimingPlot,
		NotBeforeTurn: notBefore,
		Duration:      WhileInZoneDuration(),
		CastOnly:      true,
		Source:        source,
		Label:         "Plotted — cast it without paying its mana cost",
	}, []Card{*exiled})
}

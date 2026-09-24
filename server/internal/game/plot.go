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
//	702.170a Plot is a keyword that functions while the card with plot
//	         is in a player's hand. "Plot [cost]" means "Any time you
//	         have priority during your main phase while the stack is
//	         empty, you may exile this card from your hand and pay
//	         [cost]. It becomes a plotted card."
//
// Both halves live here, and they share one function:
//
//   - the EFFECT half (#1318): a card already in exile becomes plotted
//     — Aven Interrupter's "exile target spell. It becomes plotted." —
//     PlotExiledCardForEffect;
//   - the KEYWORD half (#1342): the CR 116.2 special action that pays
//     the plot cost and exiles a card with plot from its owner's hand
//     face up — plotLocked below, the performer special_action.go's
//     verb dispatches to. It routes the card to exile and then calls
//     PlotExiledCardForEffect on what landed, so a card plotted from
//     hand and a card plotted by Aven Interrupter are the same object
//     with the same permission.
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
//	"any turn after the turn in      NotBeforeSeq — see below
//	 which it became plotted"
//
// # The floor
//
// Turn.Seq changes at every turn boundary, so the printed "any turn
// after the turn in which it became plotted" is the next sequence.

// PlotExiledCardForEffect makes a card in exile plotted: its owner may
// cast it without paying its mana cost, in their main phase with the
// stack empty, on any later turn, for as long as it stays in exile.
//
// A card that is not in exile is left alone and is not an error — the
// instruction reads "it becomes plotted" off a move that may have gone
// elsewhere (a commander taking CR 903.9's offer), and the caller that
// cares asks through ExileSpellThenForEffect's `exiled`.
//
// It emits EventBecomesPlotted once the permission is in place (#1382),
// which is what "when this card becomes plotted" watches. The plotter
// — the event's Actor — is the resolving item's controller, since an
// effect is the only thing that calls this from outside the package;
// with nothing resolving it falls back to the owner.
//
// Caller must hold g.mu.
func (g *Game) PlotExiledCardForEffect(cardID uuid.UUID, source uuid.UUID) {
	g.plotExiledCardLocked(cardID, source, g.resolutionCauseLocked().Controller)
}

// plotExiledCardLocked is PlotExiledCardForEffect with the plotter
// named. The special action names it (plotLocked): Game.resolving can
// be stale in the window between a resolution and play moving on,
// which is exactly when a special action is taken (move_cause.go), so
// inferring it there could credit the previous spell's controller.
//
// Caller must hold g.mu.
func (g *Game) plotExiledCardLocked(cardID, source, plotter uuid.UUID) {
	exiled := exiledCardByIDLocked(g, cardID)
	if exiled == nil {
		return
	}
	owner := exiled.Owner
	if plotter == uuid.Nil {
		plotter = owner
	}
	notBefore := g.Turn.Seq + 1
	g.GrantCastPermissionToCardsForEffect(CastPermission{
		Player:       owner,
		Zone:         ZoneExile,
		Cost:         "{0}",
		Timing:       TimingPlot,
		NotBeforeSeq: notBefore,
		Duration:     WhileInZoneDuration(),
		CastOnly:     true,
		Source:       source,
		Label:        "Plotted — cast it without paying its mana cost",
	}, []Card{*exiled})
	// #1382: CR 702.170c/d — the card has now become plotted. Emitted
	// AFTER the grant so a watcher sees a card that is already
	// plotted, and emitted here rather than in either caller so the
	// special action and every "it becomes plotted" effect fire it
	// alike. The card is in exile, which is where "when this card
	// becomes plotted" watches from (effects.WhenThisBecomesPlotted).
	g.EmitEvent(Event{
		Kind:   EventBecomesPlotted,
		Actor:  plotter,
		Source: source,
		CardID: cardID,
		Target: cardID,
	})
}

// plotLocked is the plot special action's performer, run by
// PerformSpecialAction once the plot cost is paid (CR 702.170a).
//
// The card goes to exile FACE UP — CR 702.170a says nothing about
// hiding it, and a plotted card is public (unlike a foretold one) —
// through the zone route with MoveCauseSpecialAction, exactly as
// suspend's does, so every entry replacement and CR 903.9 commander
// offer sees an ordinary exile. It becomes plotted only if it LANDED:
// a commander whose owner took the command zone never reached exile,
// so there is nothing to plot, and PlotExiledCardForEffect's own
// not-in-exile guard would say the same.
//
// The source of the permission is the card itself — the keyword is
// its own, not another permanent's.
//
// Caller must hold g.mu (write).
func (g *Game) plotLocked(p *Player, cardID uuid.UUID, _ SpecialAction) error {
	owner := p.ID
	return g.routeAllThenLocked(zoneRoute{
		Dst:   ZoneExile,
		Actor: owner,
		// #1320: a special action (CR 702.170a), not a spell or ability.
		Cause: MoveCause{Kind: MoveCauseSpecialAction, Controller: owner},
	}, []uuid.UUID{cardID}, func(g *Game, landed []uuid.UUID) error {
		if len(landed) != 1 {
			return nil
		}
		g.plotExiledCardLocked(landed[0], landed[0], owner)
		return nil
	})
}

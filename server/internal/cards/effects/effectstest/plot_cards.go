package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// plot_cards_test.go — #1342, the plot keyword's proof cards against
// the REAL catalog. The special action's window and the plotted
// card's cast window are pinned in the game package (plot_keyword_test
// and plot_test); what only a catalog test can check is that each card
// file declared plot at its printed cost, and that each card still
// does what it says when it arrives by the free cast.

const (
	DjinnOfFoolsFallOracle   = "7dbb540f-ee96-4115-b6ef-28ffc73da8b7"
	SpinewoodsPaladinOracle  = "d2aa9649-8394-48ca-bee6-a044e5a86b38"
	BeastbondOutcasterOracle = "19a32e29-45ec-433e-9cb4-b2c32cac8f80"
	PlanTheHeistOracle       = "a9471b33-b5b0-408a-9900-6964e5fe42da"
)

// plotOf reads the card's declared plot action, or nil.
func PlotOf(oracle string) *game.SpecialAction {
	for _, sa := range game.SpecialActionsFor(oracle) {
		if sa.Kind == game.SpecialActionPlot {
			out := sa
			return &out
		}
	}
	return nil
}

// plotHandCard seeds a card with a real mana cost and P/T in p's hand,
// so a "free" cast is observably free and a creature survives CR 704.
func PlotHandCard(p *game.Player, name, typeLine, manaCost, oracle string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		OracleID: oracle, Owner: p.ID, Controller: p.ID,
		Power: power, Toughness: toughness,
	})
	return id
}

// toNextTurnMainOf walks through the rest of the table's turns to the
// seat's NEXT precombat main phase — a real later turn, not a bumped
// counter.
func ToNextTurnMainOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	AdvanceToMainOf(t, g, seat)
}

// plotAndCastLater plots the card in the active seat's main phase,
// checks the same-turn cast is refused, walks to that seat's next
// turn and casts it from exile with an EMPTY pool.
func PlotAndCastLater(t *testing.T, g *game.Game, id uuid.UUID, plotMana ...game.ManaToken) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	p := g.Seats[seat]
	AdvanceToMainOf(t, g, seat)
	p.ManaPool.AddMana(plotMana...)
	if err := g.PerformSpecialAction(p.ID, id, game.SpecialActionPlot, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("plot: %v", err)
	}
	if len(p.ManaPool) != 0 {
		t.Fatalf("mana pool: %d tokens left after plotting — wrong plot cost", len(p.ManaPool))
	}
	if !g.Exile.Contains(id) {
		t.Fatal("the plotted card is not in exile")
	}
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Fatal("cast a plotted card on the turn it was plotted (CR 702.170d)")
	}
	ToNextTurnMainOf(t, g, seat)
	p.ManaPool = nil
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("free cast of the plotted card on a later turn: %v", err)
	}
	PassPriorityAroundTable(t, g)
}

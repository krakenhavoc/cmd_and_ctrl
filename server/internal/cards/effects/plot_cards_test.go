package effects

import (
	"slices"
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
	djinnOfFoolsFallOracle   = "7dbb540f-ee96-4115-b6ef-28ffc73da8b7"
	spinewoodsPaladinOracle  = "d2aa9649-8394-48ca-bee6-a044e5a86b38"
	beastbondOutcasterOracle = "19a32e29-45ec-433e-9cb4-b2c32cac8f80"
	planTheHeistOracle       = "a9471b33-b5b0-408a-9900-6964e5fe42da"
)

// plotOf reads the card's declared plot action, or nil.
func plotOf(oracle string) *game.SpecialAction {
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
func plotHandCard(p *game.Player, name, typeLine, manaCost, oracle string, power, toughness int) uuid.UUID {
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
func toNextTurnMainOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	advanceToMainOf(t, g, seat)
}

// plotAndCastLater plots the card in the active seat's main phase,
// checks the same-turn cast is refused, walks to that seat's next
// turn and casts it from exile with an EMPTY pool.
func plotAndCastLater(t *testing.T, g *game.Game, id uuid.UUID, plotMana ...game.ManaToken) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	p := g.Seats[seat]
	advanceToMainOf(t, g, seat)
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
	toNextTurnMainOf(t, g, seat)
	p.ManaPool = nil
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("free cast of the plotted card on a later turn: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// Each card declares plot at the cost its oracle text prints. The
// later cast is free, so there is no cast cost on the declaration.
func TestPlotCardsDeclareTheirPrintedCost(t *testing.T) {
	for _, tc := range []struct {
		name   string
		oracle string
		cost   string
	}{
		{"Djinn of Fool's Fall", djinnOfFoolsFallOracle, "{3}{U}"},
		{"Spinewoods Paladin", spinewoodsPaladinOracle, "{3}{G}"},
		{"Beastbond Outcaster", beastbondOutcasterOracle, "{1}{G}"},
		{"Plan the Heist", planTheHeistOracle, "{3}{U}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sa := plotOf(tc.oracle)
			if sa == nil {
				t.Fatalf("%s declares no plot special action", tc.name)
			}
			if sa.Cost != tc.cost {
				t.Errorf("plot cost: got %q, want %q", sa.Cost, tc.cost)
			}
			if sa.CastCost != "" {
				t.Errorf("plot cast cost: got %q, want none — the plotted cast is free (CR 702.170d)", sa.CastCost)
			}
			if sa.Label != "Plot "+tc.cost {
				t.Errorf("label: got %q, want %q", sa.Label, "Plot "+tc.cost)
			}
		})
	}
}

// Djinn of Fool's Fall end to end: plot for {3}{U}, not castable this
// turn, cast free on the seat's next turn, and it lands a flier.
func TestDjinnOfFoolsFallPlotsAndCastsFreeOnALaterTurn(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[g.Turn.ActiveSeat]
	id := plotHandCard(p, "Djinn of Fool's Fall", "Creature — Djinn", "{4}{U}", djinnOfFoolsFallOracle, 4, 3)
	plotAndCastLater(t, g, id,
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "U"})
	if !g.Battlefield.Contains(id) {
		t.Fatal("the plotted Djinn never reached the battlefield")
	}
	if !slices.Contains(effectiveAbilities(t, g, id), "flying") {
		t.Error("the Djinn has no flying")
	}
}

// Spinewoods Paladin: a creature cast free from exile still enters,
// so its enters trigger still gains 3 life.
func TestSpinewoodsPaladinPlottedStillGainsThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[g.Turn.ActiveSeat]
	id := plotHandCard(p, "Spinewoods Paladin", "Creature — Human Knight", "{4}{G}", spinewoodsPaladinOracle, 5, 4)
	before := p.Life
	plotAndCastLater(t, g, id,
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "G"})
	if !g.Battlefield.Contains(id) {
		t.Fatal("the plotted Paladin never reached the battlefield")
	}
	if got := p.Life; got != before+3 {
		t.Errorf("life: got %d, want %d", got, before+3)
	}
	if !slices.Contains(effectiveAbilities(t, g, id), "trample") {
		t.Error("the Paladin has no trample")
	}
}

// Beastbond Outcaster: the intervening-if reads current power. With
// no creature of power 4 or greater there is no card; with one there
// is exactly one.
func TestBeastbondOutcasterDrawsOnlyWithAPowerFourCreature(t *testing.T) {
	for _, tc := range []struct {
		name  string
		big   bool
		drawn int
	}{
		{"no power-4 creature", false, 0},
		{"a power-4 creature", true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			p := g.Seats[g.Turn.ActiveSeat]
			if tc.big {
				pushBattlefieldCardWithTimestamp(g, game.Card{
					InstanceID: uuid.New(), Name: "Big Beast", TypeLine: "Creature — Beast",
					Power: 4, Toughness: 4, Owner: p.ID, Controller: p.ID,
				})
			}
			advanceToMainOf(t, g, g.Turn.ActiveSeat)
			id := plotHandCard(p, "Beastbond Outcaster", "Creature — Human Druid", "", beastbondOutcasterOracle, 3, 3)
			before := p.Hand.Size()
			if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast Beastbond Outcaster: %v", err)
			}
			passPriorityAroundTable(t, g)
			// The cast took the Outcaster out of hand; anything beyond
			// that is the draw.
			if got := p.Hand.Size() - (before - 1); got != tc.drawn {
				t.Errorf("cards drawn: got %d, want %d", got, tc.drawn)
			}
			// CR 603.4: the condition is checked as the creature
			// enters too, so with no big creature the trigger never
			// goes on the stack at all — not merely does nothing.
			triggered := 0
			for _, ev := range g.Events {
				if ev.Kind == game.EventTrigger && ev.Source == id {
					triggered++
				}
			}
			if triggered != tc.drawn {
				t.Errorf("enters triggers: got %d, want %d", triggered, tc.drawn)
			}
		})
	}
}

// Plan the Heist: "Surveil 3 if you have no cards in hand. Then draw
// three cards." With another card in hand, no surveil; with the spell
// the only card, it surveils first — and the spell itself, on the
// stack, does not count against its own condition.
func TestPlanTheHeistSurveilsOnlyWithAnEmptyHand(t *testing.T) {
	t.Run("a card in hand", func(t *testing.T) {
		g := newCatalogGame(t)
		p := g.Seats[g.Turn.ActiveSeat]
		castCatalogSpell(t, g, "Plan the Heist", "Sorcery", planTheHeistOracle, nil)
		before := p.Hand.Size()
		if before == 0 {
			t.Fatal("the fixture's hand is empty; this case needs a card in it")
		}
		passPriorityAroundTable(t, g)
		if c := latestChoiceOfKind(g, game.PendingChoiceSurveil); c != nil {
			t.Fatal("Plan the Heist surveilled with a card in hand")
		}
		if got := p.Hand.Size(); got != before+3 {
			t.Errorf("hand: got %d, want %d", got, before+3)
		}
	})
	t.Run("an empty hand", func(t *testing.T) {
		g := newCatalogGame(t)
		p := g.Seats[g.Turn.ActiveSeat]
		p.Hand.Cards = nil
		castCatalogSpell(t, g, "Plan the Heist", "Sorcery", planTheHeistOracle, nil)
		passPriorityAroundTable(t, g)
		c := latestChoiceOfKind(g, game.PendingChoiceSurveil)
		if c == nil {
			t.Fatal("Plan the Heist did not surveil with an empty hand")
		}
		if p.Hand.Size() != 0 {
			t.Fatalf("drew before the surveil was answered: hand %d", p.Hand.Size())
		}
		if err := g.ResolveSurveil(c.ID, p.ID, nil, c.ScryCards); err != nil {
			t.Fatalf("ResolveSurveil: %v", err)
		}
		if got := p.Hand.Size(); got != 3 {
			t.Errorf("hand after surveil: got %d, want 3", got)
		}
	})
}

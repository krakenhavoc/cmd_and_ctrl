package testsp

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Each card declares plot at the cost its oracle text prints. The
// later cast is free, so there is no cast cost on the declaration.
func TestPlotCardsDeclareTheirPrintedCost(t *testing.T) {
	for _, tc := range []struct {
		name   string
		oracle string
		cost   string
	}{
		{"Djinn of Fool's Fall", effectstest.DjinnOfFoolsFallOracle, "{3}{U}"},
		{"Spinewoods Paladin", effectstest.SpinewoodsPaladinOracle, "{3}{G}"},
		{"Beastbond Outcaster", effectstest.BeastbondOutcasterOracle, "{1}{G}"},
		{"Plan the Heist", effectstest.PlanTheHeistOracle, "{3}{U}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sa := effectstest.PlotOf(tc.oracle)
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
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PlotHandCard(p, "Djinn of Fool's Fall", "Creature — Djinn", "{4}{U}", effectstest.DjinnOfFoolsFallOracle, 4, 3)
	effectstest.PlotAndCastLater(t, g, id,
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "U"})
	if !g.Battlefield.Contains(id) {
		t.Fatal("the plotted Djinn never reached the battlefield")
	}
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, id), "flying") {
		t.Error("the Djinn has no flying")
	}
}

// Spinewoods Paladin: a creature cast free from exile still enters,
// so its enters trigger still gains 3 life.
func TestSpinewoodsPaladinPlottedStillGainsThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PlotHandCard(p, "Spinewoods Paladin", "Creature — Human Knight", "{4}{G}", effectstest.SpinewoodsPaladinOracle, 5, 4)
	before := p.Life
	effectstest.PlotAndCastLater(t, g, id,
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "G"})
	if !g.Battlefield.Contains(id) {
		t.Fatal("the plotted Paladin never reached the battlefield")
	}
	if got := p.Life; got != before+3 {
		t.Errorf("life: got %d, want %d", got, before+3)
	}
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, id), "trample") {
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
			g := effectstest.NewCatalogGame(t)
			p := g.Seats[g.Turn.ActiveSeat]
			if tc.big {
				effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
					InstanceID: uuid.New(), Name: "Big Beast", TypeLine: "Creature — Beast",
					Power: 4, Toughness: 4, Owner: p.ID, Controller: p.ID,
				})
			}
			effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
			id := effectstest.PlotHandCard(p, "Beastbond Outcaster", "Creature — Human Druid", "", effectstest.BeastbondOutcasterOracle, 3, 3)
			before := p.Hand.Size()
			if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast Beastbond Outcaster: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
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
		g := effectstest.NewCatalogGame(t)
		p := g.Seats[g.Turn.ActiveSeat]
		effectstest.CastCatalogSpell(t, g, "Plan the Heist", "Sorcery", effectstest.PlanTheHeistOracle, nil)
		before := p.Hand.Size()
		if before == 0 {
			t.Fatal("the fixture's hand is empty; this case needs a card in it")
		}
		effectstest.PassPriorityAroundTable(t, g)
		if c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceSurveil); c != nil {
			t.Fatal("Plan the Heist surveilled with a card in hand")
		}
		if got := p.Hand.Size(); got != before+3 {
			t.Errorf("hand: got %d, want %d", got, before+3)
		}
	})
	t.Run("an empty hand", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		p := g.Seats[g.Turn.ActiveSeat]
		p.Hand.Cards = nil
		effectstest.CastCatalogSpell(t, g, "Plan the Heist", "Sorcery", effectstest.PlanTheHeistOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceSurveil)
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

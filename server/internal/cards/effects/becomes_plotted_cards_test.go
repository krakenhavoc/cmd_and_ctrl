package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// becomes_plotted_cards_test.go — #1382: "when this card becomes
// plotted" against the REAL catalog. The event is pinned in the game
// package (plot_test.go); what only a catalog test can show is that
// each proof card's trigger fires from exile on both routes a card
// becomes plotted by — the hand special action and an "it becomes
// plotted" effect — and never on a plain exile.

const (
	longhornSharpshooterOracle = "6a0a7b02-10e6-4dbf-8356-659095519480"
	aloeAlchemistOracle        = "97489ef7-98c3-4700-bbe6-185215d41b25"
)

// plotFromHand takes the plot special action for the active seat in
// its main phase, paying from a pool seeded with `mana`.
func plotFromHand(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, mana ...game.ManaToken) {
	t.Helper()
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	p.ManaPool.AddMana(mana...)
	if err := g.PerformSpecialAction(p.ID, id, game.SpecialActionPlot, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("plot: %v", err)
	}
	if !g.Exile.Contains(id) {
		t.Fatal("the plotted card is not in exile")
	}
}

// plottedTriggersFrom counts the EventTrigger events whose source is
// the plotted card — a trigger that fired at all, whatever it then did.
func plottedTriggersFrom(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventTrigger && ev.Source == id {
			n++
		}
	}
	return n
}

// Longhorn Sharpshooter plotted from hand: the trigger fires from
// exile, its owner picks the target, and the opponent takes 2.
func TestLonghornSharpshooterPlottedFromHandDealsTwo(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := plotHandCard(p, "Longhorn Sharpshooter", "Creature — Minotaur Rogue", "{2}{R}", longhornSharpshooterOracle, 3, 3)
	before := opp.Life

	plotFromHand(t, g, p, id,
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "R"})
	if len(p.ManaPool) != 0 {
		t.Fatalf("mana pool: %d tokens left — wrong plot cost", len(p.ManaPool))
	}
	pick := latestPickTarget(g, p.ID)
	if pick == nil {
		t.Fatal("no target prompt for the owner after plotting — the becomes-plotted trigger did not fire")
	}
	if err := g.ResolvePickTarget(pick.ID, p.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := opp.Life; got != before-2 {
		t.Errorf("opponent life: got %d, want %d", got, before-2)
	}
	if !g.Exile.Contains(id) {
		t.Error("the Sharpshooter left exile — it stays plotted until cast")
	}
	if n := plottedTriggersFrom(g, id); n != 1 {
		t.Errorf("becomes-plotted triggers: got %d, want 1", n)
	}

	// "THIS card": another card becoming plotted while the
	// Sharpshooter sits plotted in exile does not fire it again.
	djinn := plotHandCard(p, "Djinn of Fool's Fall", "Creature — Djinn", "{4}{U}", djinnOfFoolsFallOracle, 4, 3)
	plotFromHand(t, g, p, djinn,
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "U"})
	if pick := latestPickTarget(g, p.ID); pick != nil {
		t.Fatalf("plotting ANOTHER card fired a target prompt from source %s", pick.Source)
	}
	if len(g.PendingTriggers) != 0 {
		t.Fatalf("plotting ANOTHER card left %d pending triggers", len(g.PendingTriggers))
	}
	passPriorityAroundTable(t, g)
	if n := plottedTriggersFrom(g, id); n != 1 {
		t.Errorf("becomes-plotted triggers after ANOTHER card was plotted: got %d, want 1", n)
	}
}

// Aloe Alchemist plotted from hand: +3/+2 and trample on the target
// until end of turn.
func TestAloeAlchemistPlottedFromHandPumpsTarget(t *testing.T) {
	g := newCatalogGame(t)
	p := g.Seats[g.Turn.ActiveSeat]
	bear := uuid.New()
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: bear, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	id := plotHandCard(p, "Aloe Alchemist", "Creature — Plant Warlock", "{1}{G}", aloeAlchemistOracle, 3, 2)

	plotFromHand(t, g, p, id, game.ManaToken{Color: "C"}, game.ManaToken{Color: "G"})
	pickCard(t, g, p.ID, bear)
	passPriorityAroundTable(t, g)

	if got, want := effectivePower(t, g, bear), 5; got != want {
		t.Errorf("power: got %d, want %d", got, want)
	}
	if got, want := effectiveToughness(t, g, bear), 4; got != want {
		t.Errorf("toughness: got %d, want %d", got, want)
	}
	if !slices.Contains(effectiveAbilities(t, g, bear), "trample") {
		t.Error("the target did not gain trample")
	}
}

// An opponent's Aven Interrupter plots the Sharpshooter off the stack:
// "this card becomes plotted" is true whoever plotted it, and the
// trigger belongs to the card's OWNER (CR 108.4), who aims it.
func TestLonghornSharpshooterPlottedByAnOpponentsAvenInterrupter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	// The opponent casts the Sharpshooter in their own main phase, and
	// I hold priority over it with a flashed-in Aven Interrupter.
	advanceToMainOf(t, g, 1)
	longhorn := plotHandCard(opp, "Longhorn Sharpshooter", "Creature — Minotaur Rogue", "", longhornSharpshooterOracle, 3, 3)
	if err := g.CastSpell(opp.ID, longhorn, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent casts the Sharpshooter: %v", err)
	}
	g.WithWriteLock(func() { g.Turn.PriorityHolder = 0 })
	aven := plotHandCard(me, "Aven Interrupter", "Creature — Bird Rogue", "", avenInterrupterOracle, 2, 2)
	if err := g.CastSpell(me.ID, aven, game.CastSpellParams{}); err != nil {
		t.Fatalf("flash in Aven Interrupter: %v", err)
	}
	passUntilOnBattlefield(t, g, aven)
	pickCard(t, g, me.ID, longhorn)
	myLife := me.Life
	for i := 0; i < 8 && latestPickTarget(g, opp.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Exile.Contains(longhorn) {
		t.Fatal("Aven Interrupter did not exile the Sharpshooter")
	}
	// The event names the PLOTTER — Aven's controller, from the
	// resolving trigger — and Aven as the source; the trigger it fires
	// is still the owner's.
	var plotted []game.Event
	for _, ev := range g.Events {
		if ev.Kind == game.EventBecomesPlotted {
			plotted = append(plotted, ev)
		}
	}
	if len(plotted) != 1 || plotted[0].CardID != longhorn || plotted[0].Actor != me.ID || plotted[0].Source != aven {
		t.Fatalf("EventBecomesPlotted = %+v, want one naming the Sharpshooter, me as plotter and Aven as source", plotted)
	}
	pick := latestPickTarget(g, opp.ID)
	if pick == nil {
		t.Fatal("no target prompt for the Sharpshooter's OWNER after Aven Interrupter plotted it")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("the plotter was asked to aim the Sharpshooter's trigger; its owner controls it")
	}
	if err := g.ResolvePickTarget(pick.ID, opp.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Life; got != myLife-2 {
		t.Errorf("my life: got %d, want %d", got, myLife-2)
	}
}

// A plain exile is not plotting: a Sharpshooter exiled from the stack
// with nothing making it plotted fires nothing, and neither does one
// moved from hand to exile by an ordinary route.
func TestLonghornSharpshooterPlainExileDoesNotTrigger(t *testing.T) {
	t.Run("exiled from the stack", func(t *testing.T) {
		g := newCatalogGame(t)
		id := castCatalogSpell(t, g, "Longhorn Sharpshooter", "Creature — Minotaur Rogue", longhornSharpshooterOracle, nil)
		g.WithWriteLock(func() {
			if err := g.ExileSpellThenForEffect(id, nil); err != nil {
				t.Fatalf("exile: %v", err)
			}
		})
		assertNotPlottedAndSilent(t, g, id)
	})
	t.Run("exiled from hand", func(t *testing.T) {
		g := newCatalogGame(t)
		p := g.Seats[g.Turn.ActiveSeat]
		id := plotHandCard(p, "Longhorn Sharpshooter", "Creature — Minotaur Rogue", "{2}{R}", longhornSharpshooterOracle, 3, 3)
		g.WithWriteLock(func() {
			if err := g.ExileCardForEffect(id); err != nil {
				t.Fatalf("exile: %v", err)
			}
		})
		assertNotPlottedAndSilent(t, g, id)
	})
}

func assertNotPlottedAndSilent(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	if !g.Exile.Contains(id) {
		t.Fatal("the card is not in exile")
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventBecomesPlotted {
			t.Fatal("a plain exile emitted EventBecomesPlotted")
		}
	}
	passPriorityAroundTable(t, g)
	if n := plottedTriggersFrom(g, id); n != 0 {
		t.Errorf("becomes-plotted triggers after a plain exile: got %d, want 0", n)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatalf("a plain exile queued a target prompt from source %s", c.Source)
		}
	}
}

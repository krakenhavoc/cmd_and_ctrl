package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// free_cast_command_test.go — #6242: "you may cast your commander from
// the command zone without paying its mana cost" (Geode Golem). CR 903.8:
// the commander tax still applies to the alternative cost (CR 118.9), and
// the cast counts toward it.

func pushCommanderFor(g *Game, p *Player) uuid.UUID {
	c := NewCard("Test Commander", p.ID)
	c.TypeLine, c.ManaCost = "Legendary Creature — Bear", "{3}{G}{G}"
	c.Power, c.Toughness = 4, 4
	c.IsCommander = true
	p.Command.PushTop(c)
	return c.InstanceID
}

// offerFreeCommander opens a resolution for `p` and offers the free
// cast of their commanders inside it, as Geode Golem's trigger does.
func offerFreeCommander(t *testing.T, g *Game, p *Player) *PendingChoice {
	t.Helper()
	counteringFor(g, p)
	src := uuid.New()
	g.WithWriteLock(func() {
		perm := g.FreeCastFromCommandPermission(p.ID, src, "Geode Golem")
		if !g.OfferResolutionCastForEffect(ResolutionCastOffer{
			Chooser: p.ID, Source: src, Reason: "Geode Golem",
			Cards:      g.OwnCommandersInCommandZoneForEffect(p.ID),
			Permission: &perm, Keyword: MayCastKeywordFreeCast,
			AcceptLabel: "Cast it free", DeclineLabel: "Leave it in the command zone",
		}) {
			t.Fatal("the offer was not taken")
		}
	})
	return guileOffer(g, p.ID)
}

// The first cast is free: an empty pool pays it, it is a cast from the
// command zone and it counts for the tax.
func TestFreeCommanderCastPaysNothingTheFirstTime(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	cmd := pushCommanderFor(g, me)
	theirs := pushCommanderFor(g, opp)
	offer := offerFreeCommander(t, g, me)
	if offer == nil || !slices.Contains(offer.DiscardOptions, cmd) || slices.Contains(offer.DiscardOptions, theirs) {
		t.Fatal("the offer must name my commander and never an opponent's")
	}
	if err := g.CastForResolutionChoice(me.ID, offer.ID, cmd, CastSpellParams{FromZone: "command", Strict: true}); err != nil {
		t.Fatalf("free commander cast with an empty pool: %v", err)
	}
	if !g.Stack.Contains(cmd) {
		t.Fatal("the commander is not on the stack")
	}
	if n := me.CommanderCasts[cmd]; n != 1 {
		t.Errorf("CommanderCasts = %d, want 1: the free cast counts (CR 903.8)", n)
	}
	if guileOffer(g, me.ID) != nil {
		t.Error("the offer stayed open after its one cast")
	}
}

// The second time it costs the {2} of commander tax and nothing more:
// an empty pool can't pay, two mana can.
func TestFreeCommanderCastStillPaysTheTax(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	cmd := pushCommanderFor(g, me)
	me.CommanderCasts = map[uuid.UUID]int{cmd: 1}
	offer := offerFreeCommander(t, g, me)
	if offer == nil {
		t.Fatal("no offer")
	}
	if err := g.CastForResolutionChoice(me.ID, offer.ID, cmd, CastSpellParams{FromZone: "command", Strict: true}); err == nil {
		t.Fatal("a second free cast with an empty pool succeeded; the {2} tax is owed")
	}
	if !me.Command.Contains(cmd) {
		t.Fatal("the refused cast moved the commander")
	}
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
	if err := g.CastForResolutionChoice(me.ID, offer.ID, cmd, CastSpellParams{FromZone: "command", Strict: true}); err != nil {
		t.Fatalf("second free cast with {2}: %v", err)
	}
	if n := me.CommanderCasts[cmd]; n != 2 {
		t.Errorf("CommanderCasts = %d, want 2", n)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left: the tax is {2} and the mana cost is not paid", me.ManaPool.Total())
	}
}

// Without the grant a commander cast pays its mana cost as always.
func TestCommanderCastWithoutTheGrantPaysItsCost(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	cmd := pushCommanderFor(g, me)
	if err := g.CastSpell(me.ID, cmd, CastSpellParams{FromZone: "command", Strict: true}); err == nil {
		t.Fatal("a commander cast with an empty pool and no grant succeeded")
	}
	if err := g.CastSpell(me.ID, cmd, CastSpellParams{FromZone: "command", Strict: true, AlternativeCost: AltCostKeyGrantedCost}); err == nil {
		t.Fatal("a claim of the free cast with no grant succeeded")
	}
}

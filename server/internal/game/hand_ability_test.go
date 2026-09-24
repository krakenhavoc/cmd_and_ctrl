package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// hand_ability_test.go — #660 / ADR 0062: an activated ability that
// functions from the hand, and the discard components that pay for it.
//
// What is pinned here is the ENGINE half: the zone dimension on the
// activation path, the two discard components, the cost's
// indivisibility (CR 601.2h / 602.2b) including for a commander, and
// the fact that a cycling still emits both of its events. The catalog
// half — the six cycling cards, Fauna Shaman, and the constructors —
// is in cards/effects/cycling_test.go.

// cyclingAbility is "{2}, Discard this card: Draw a card" as the
// catalog's Cycling constructor builds it, spelled out here so the
// game package's tests need no catalog import.
func cyclingAbility(cost string) ActivatedAbilityShape {
	return ActivatedAbilityShape{
		Label:   "Cycling " + cost,
		Cost:    AbilityCost{Mana: cost, DiscardSelf: true},
		Zones:   []ZoneKind{ZoneHand},
		Cycling: true,
		Effect: func(g *Game, item *StackItem) error {
			return g.DrawNForEffect(item.Controller, 1)
		},
	}
}

// pushHandCard puts a card carrying `abilities` into p's hand and
// returns its instance ID.
func pushHandCard(g *Game, p *Player, name string, abilities ...ActivatedAbilityShape) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = "Instant"
	c.ActivatedAbilities = abilities
	p.Hand.PushTop(c)
	return c.InstanceID
}

// cycleWatcher records the two events a cycling owes.
type cycleWatcher struct {
	discards []Event
	cycles   []Event
}

func (w *cycleWatcher) OnEvent(_ *Game, ev Event) {
	switch ev.Kind {
	case EventDiscardCard:
		w.discards = append(w.discards, ev)
	case EventCycle:
		w.cycles = append(w.cycles, ev)
	}
}

// Cycling from hand: the mana and the discard are paid, the card is in
// the graveyard, both events fired, the ability is on the stack, and it
// resolves to a draw.
func TestCycleFromHandPaysDiscardsAndDraws(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	w := &cycleWatcher{}
	g.RegisterListener(w)
	card := pushHandCard(g, me, "Ketria Triome", cyclingAbility("{2}"))
	me.ManaPool.AddMana(ManaToken{Color: "C"})
	me.ManaPool.AddMana(ManaToken{Color: "C"})
	handBefore := len(me.Hand.Cards)

	if err := g.ActivateCatalogAbility(me.ID, card, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want the {2} spent", me.ManaPool)
	}
	if me.Hand.Contains(card) {
		t.Error("the cycled card is still in hand")
	}
	if !me.Graveyard.Contains(card) {
		t.Error("the cycled card is not in the graveyard")
	}
	if len(w.discards) != 1 || w.discards[0].CardID != card || w.discards[0].Actor != me.ID {
		t.Errorf("discards = %+v, want one for %v by %v", w.discards, card, me.ID)
	}
	// CR 702.29b: activating the ability IS cycling the card, so the
	// event fires at announce — and after the discard, because
	// CR 702.29c puts the watchers' view of the card in the
	// graveyard.
	if len(w.cycles) != 1 || w.cycles[0].CardID != card || w.cycles[0].Actor != me.ID {
		t.Fatalf("cycles = %+v, want one for %v by %v", w.cycles, card, me.ID)
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("StackMeta has %d items, want the cycling ability", len(g.StackMeta))
	}

	passBothForTest(g)
	// One card left the hand as a cost, one arrived from the draw.
	if got := len(me.Hand.Cards); got != handBefore {
		t.Errorf("hand = %d cards, want %d (discarded one, drew one)", got, handBefore)
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("StackMeta has %d items after resolution, want none", len(g.StackMeta))
	}
}

// CR 113.6: the zone dimension cuts both ways. A cycling ability is
// not activatable from the battlefield or the graveyard, and an
// ordinary battlefield ability is not activatable from hand. Both
// refuse with the same error and pay nothing.
func TestAbilityZoneDimensionRefusesTheWrongZone(t *testing.T) {
	t.Run("cycling from the battlefield", func(t *testing.T) {
		g := newActiveGame(t)
		advanceTo(t, g, StepPrecombatMain)
		me := g.Seats[0]
		c := NewCard("Ketria Triome", me.ID)
		c.TypeLine = "Land"
		c.Controller = me.ID
		c.ActivatedAbilities = []ActivatedAbilityShape{cyclingAbility("{2}")}
		g.Battlefield.PushTop(c)
		me.ManaPool.AddMana(ManaToken{Color: "C"})
		me.ManaPool.AddMana(ManaToken{Color: "C"})

		err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, ActivateAbilityParams{})
		if !errors.Is(err, ErrActivationZoneNotAllowed) {
			t.Fatalf("err = %v, want ErrActivationZoneNotAllowed", err)
		}
		if len(me.ManaPool) != 2 {
			t.Errorf("pool = %v, want the {2} unspent", me.ManaPool)
		}
		if len(g.StackMeta) != 0 {
			t.Error("a refused activation announced an ability")
		}
	})

	t.Run("cycling from the graveyard", func(t *testing.T) {
		g := newActiveGame(t)
		advanceTo(t, g, StepPrecombatMain)
		me := g.Seats[0]
		c := NewCard("Ketria Triome", me.ID)
		c.ActivatedAbilities = []ActivatedAbilityShape{cyclingAbility("{2}")}
		me.Graveyard.PushTop(c)
		me.ManaPool.AddMana(ManaToken{Color: "C"})
		me.ManaPool.AddMana(ManaToken{Color: "C"})

		if err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, ActivateAbilityParams{}); !errors.Is(err, ErrActivationZoneNotAllowed) {
			t.Fatalf("err = %v, want ErrActivationZoneNotAllowed", err)
		}
		if !me.Graveyard.Contains(c.InstanceID) {
			t.Error("the card moved")
		}
	})

	t.Run("a battlefield ability from hand", func(t *testing.T) {
		g := newActiveGame(t)
		advanceTo(t, g, StepPrecombatMain)
		me := g.Seats[0]
		id := pushHandCard(g, me, "Sol Ring", ActivatedAbilityShape{
			Label:  "{2}: Draw a card.",
			Cost:   AbilityCost{Mana: "{2}"},
			Effect: func(g *Game, item *StackItem) error { return g.DrawNForEffect(item.Controller, 1) },
		})
		me.ManaPool.AddMana(ManaToken{Color: "C"})
		me.ManaPool.AddMana(ManaToken{Color: "C"})

		if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{}); !errors.Is(err, ErrActivationZoneNotAllowed) {
			t.Fatalf("err = %v, want ErrActivationZoneNotAllowed", err)
		}
	})
}

// Only the hand's owner may activate a hand ability: a card off the
// battlefield has no controller (CR 108.4), so the owner is the "you".
func TestHandAbilityIsTheOwnersAlone(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, them := g.Seats[0], g.Seats[1]
	card := pushHandCard(g, me, "Ketria Triome", cyclingAbility("{2}"))
	them.ManaPool.AddMana(ManaToken{Color: "C"})
	them.ManaPool.AddMana(ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(them.ID, card, 0, ActivateAbilityParams{}); !errors.Is(err, ErrCardCallerMismatch) {
		t.Fatalf("err = %v, want ErrCardCallerMismatch", err)
	}
	if !me.Hand.Contains(card) {
		t.Error("the card left its owner's hand")
	}
}

// CR 601.2h / 602.2b: a cost is one indivisible step, so the cost
// discard sets MustSettleNow and cannot pause. Since #1397 the CR 903.9
// question is asked BEFORE the cycling is paid for, so a declined
// commander goes to the graveyard and no prompt is left behind once the
// ability is on the stack. ADR 0062 Decision 3, ADR 0013 §5af.
func TestCyclingACommanderFromHandGoesToTheGraveyardWhenDeclined(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	cmdr := seatCommander(t, me.Hand, me)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == cmdr {
				me.Hand.Cards[i].ActivatedAbilities = []ActivatedAbilityShape{cyclingAbility("{2}")}
			}
		}
	})
	me.ManaPool.AddMana(ManaToken{Color: "C"})
	me.ManaPool.AddMana(ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, cmdr, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("cycle the commander: %v", err)
	}
	prompt := expectCommanderPrompt(t, g, me)
	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("a cost discard queued %d prompt(s) after the answer; CR 601.2h pays costs without asking", len(g.PendingChoices))
	}
	if !me.Graveyard.Contains(cmdr) {
		t.Error("the cycled commander is not in the graveyard")
	}
	if me.Command.Contains(cmdr) {
		t.Error("the cycled commander went to the command zone after its owner declined")
	}
}

// discardOutletAbility is "{T}, Discard N <clause>: Draw a card" on a
// battlefield permanent — the general component, on the side of the
// seam Fauna Shaman sits on.
func discardOutletAbility(n int, label string, match func(Card) bool) ActivatedAbilityShape {
	return ActivatedAbilityShape{
		Label: "Discard, draw",
		Cost: AbilityCost{
			Tap:          true,
			DiscardCards: &DiscardCost{N: n, Label: label, Match: match},
		},
		Effect: func(g *Game, item *StackItem) error {
			return g.DrawNForEffect(item.Controller, 1)
		},
	}
}

func pushDiscardOutlet(g *Game, p *Player, ab ActivatedAbilityShape) uuid.UUID {
	c := NewCard("Fauna Shaman", p.ID)
	c.TypeLine = "Artifact"
	c.Controller = p.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{ab}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// The general "Discard N cards" component: exactly N ids, each in the
// activator's hand, each matching the clause, and the discard is paid
// before the ability reaches the stack.
func TestDiscardCardsCostValidatesAndPaysAtAnnounce(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	bear := pushHandCard(g, me, "Bear")
	wolf := pushHandCard(g, me, "Wolf")
	bolt := pushHandCard(g, me, "Bolt")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID != bolt {
				me.Hand.Cards[i].TypeLine = "Creature — Beast"
			}
		}
	})
	src := pushDiscardOutlet(g, me, discardOutletAbility(1, "a creature card", func(c Card) bool { return c.IsCreature() }))

	// Wrong count.
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{DiscardIDs: []uuid.UUID{bear, wolf}}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("two ids for a one-card clause: err = %v, want ErrInvalidParam", err)
	}
	// No ids at all.
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("no ids: err = %v, want ErrInvalidParam", err)
	}
	// A card the clause does not admit.
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{DiscardIDs: []uuid.UUID{bolt}}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("a non-creature for \"a creature card\": err = %v, want ErrInvalidParam", err)
	}
	// A card that is not in hand.
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{DiscardIDs: []uuid.UUID{src}}); !errors.Is(err, ErrCardNotFound) {
		t.Fatalf("a battlefield card: err = %v, want ErrCardNotFound", err)
	}
	if len(me.Hand.Cards) != 3 || tappedOf(g, src) {
		t.Fatal("a refused activation paid part of the cost")
	}

	// The legal payment: discarded at announce, before the ability is
	// on the stack, so a discard payoff would see it first.
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{DiscardIDs: []uuid.UUID{wolf}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Contains(wolf) || !me.Graveyard.Contains(wolf) {
		t.Error("the named card was not discarded")
	}
	if !me.Hand.Contains(bear) || !me.Hand.Contains(bolt) {
		t.Error("the discard took a card the activator did not name")
	}
	if len(g.StackMeta) != 1 {
		t.Errorf("StackMeta has %d items, want the ability", len(g.StackMeta))
	}
}

// DiscardCostOptionsForEffect is the set the view ships and the
// enumerator solves from: hand cards matching the clause, source
// excluded. Exactly N options is the "no prompt needed" case.
func TestDiscardCostOptions(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	a := pushHandCard(g, me, "Bear")
	b := pushHandCard(g, me, "Wolf")
	bolt := pushHandCard(g, me, "Bolt")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID != bolt {
				me.Hand.Cards[i].TypeLine = "Creature — Beast"
			}
		}
	})

	creatures := &DiscardCost{N: 1, Label: "a creature card", Match: func(c Card) bool { return c.IsCreature() }}
	var got []uuid.UUID
	g.ReadSnapshot(func() { got = g.DiscardCostOptionsForEffect(me.ID, uuid.Nil, creatures) })
	if len(got) != 2 || got[0] != a || got[1] != b {
		t.Errorf("options = %v, want the two creature cards %v %v in hand order", got, a, b)
	}

	// A clause with no predicate takes anything, and the source is
	// still excluded — an ability activated FROM hand cannot pay for
	// itself with itself.
	anyCard := &DiscardCost{N: 1, Label: "a card"}
	g.ReadSnapshot(func() { got = g.DiscardCostOptionsForEffect(me.ID, a, anyCard) })
	if len(got) != 2 {
		t.Errorf("options = %v, want the two cards that are not the source", got)
	}
	for _, id := range got {
		if id == a {
			t.Error("the source is offered as its own payment")
		}
	}
}

// Undo across a cost discard: the clone taken before the activation is
// an independent game with the card still in hand, which is what the
// room's undo stack restores. The cost is atomic, so there is no
// half-undone state to land in.
func TestUndoAcrossACyclingCost(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	card := pushHandCard(g, me, "Ketria Triome", cyclingAbility("{2}"))
	me.ManaPool.AddMana(ManaToken{Color: "C"})
	me.ManaPool.AddMana(ManaToken{Color: "C"})

	before := g.Clone()
	if err := g.ActivateCatalogAbility(me.ID, card, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	if !me.Graveyard.Contains(card) {
		t.Fatal("the cost did not discard")
	}

	undone := before.Seats[0]
	if !undone.Hand.Contains(card) {
		t.Error("undo lost the cycled card from hand")
	}
	if undone.Graveyard.Contains(card) {
		t.Error("undo kept the card in the graveyard")
	}
	if len(undone.ManaPool) != 2 {
		t.Errorf("undone pool = %v, want the {2} back", undone.ManaPool)
	}
	if len(before.StackMeta) != 0 {
		t.Error("undo kept the ability on the stack")
	}
}

package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// face_down_cast_test.go — ADR 0082 / #1194: casting a card face down
// (CR 708.4) and turning it face up (CR 116.2g, CR 708.6).
//
// ADR 0069's object model is tested next door in face_down_test.go;
// what belongs here is the two MECHANICS that reach it:
//
//	1. the cast — an alternative cost that stamps the object
//	2. the resolution — the permanent enters face down, silently
//	3. the special action — the offer, the timing, the payment
//	4. turning face up — same object, real card back, trigger fires
//	5. megamorph's counter
//	6. the ways there is NO way up (CR 701.40b, CR 708.7)

const (
	morphOracle = "oracle-morph-test"
	morphCost   = "{1}{U}"
)

// morphOffer is the catalog declaration effects.Morph produces, kept
// here as data so the game package can test the shape without
// importing the catalog.
func morphOffer(kind FaceDownKind, faceUp string, counter bool) AlternativeCost {
	return AlternativeCost{
		Key:      "morph",
		Label:    "Morph — cast face down {3}",
		ManaCost: "{3}",
		FaceDown: &FaceDownCast{Kind: kind, FaceUpCost: faceUp, FaceUpCounter: counter},
	}
}

// morphGame seeds a two-seat game with one morph creature in seat 0's
// hand, the catalog offer wired, and enough mana in the pool to cast
// it face down and turn it up again.
func morphGame(t *testing.T, offer AlternativeCost) (*Game, *Player, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	me := g.Seats[0]
	c := NewCard("Willbender", me.ID)
	c.OracleID = morphOracle
	c.TypeLine = "Creature — Human Wizard"
	c.ManaCost = "{1}{U}"
	c.Colors = []string{"U"}
	c.Power, c.Toughness = 1, 2
	c.Keywords = []string{"flying"}
	id := c.InstanceID
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	withCatalogAlternativeCosts(t, altCostFor(morphOracle, offer))
	fillPool(g, me, 8)
	return g, me, id
}

// fillPool drops n colourless mana into a seat's pool. Colourless
// pays generic, and every cost in this file is generic or blue —
// the blue half is topped up by the caller where it matters.
func fillPool(g *Game, p *Player, n int) {
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			p.ManaPool.AddMana(ManaToken{Color: "U"})
		}
	})
}

// castFaceDown casts the card claiming the morph offer, from hand,
// at sorcery speed on its controller's main phase.
func castFaceDown(t *testing.T, g *Game, p *Player, id uuid.UUID, key string) error {
	t.Helper()
	g.WithWriteLock(func() {
		g.Turn.Step = StepPrecombatMain
	})
	return g.CastSpell(p.ID, id, CastSpellParams{
		FromZone:        "hand",
		AlternativeCost: key,
		Strict:          true,
	})
}

// --- 1. the cast ------------------------------------------------------

// TestCastFaceDownPutsANamelessTwoTwoOnTheStack is CR 708.4: the
// spell on the stack is a 2/2 creature spell with no name, no types
// beyond Creature, no mana cost and no text — and only its controller
// may look at it (CR 708.5).
func TestCastFaceDownPutsANamelessTwoTwoOnTheStack(t *testing.T) {
	g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
	opp := g.Seats[1]
	if err := castFaceDown(t, g, me, id, "morph"); err != nil {
		t.Fatalf("cast face down: %v", err)
	}
	c, zone, ok := cardAnywhere(g, id)
	if !ok || zone != ZoneStack {
		t.Fatalf("card is in %q, want the stack", zone)
	}
	if !c.FaceDown || c.FaceDownKind != FaceDownMorphed {
		t.Fatalf("face-down state = (%v, %q), want (true, %q)", c.FaceDown, c.FaceDownKind, FaceDownMorphed)
	}
	// CR 708.2: the object is a 2/2 creature with no name and no
	// text, whatever the card underneath says (a 1/2 flier here).
	if !c.IsCreature() {
		t.Error("a face-down spell is not a creature spell")
	}
	eff := c.printedCharacteristic()
	if eff.Power != 2 || eff.Toughness != 2 {
		t.Errorf("projected body = %d/%d, want 2/2", eff.Power, eff.Toughness)
	}
	if eff.Name != "" {
		t.Errorf("projected name = %q, want none (CR 708.2)", eff.Name)
	}
	if key := CatalogKey(c); key != "" {
		t.Errorf("CatalogKey = %q, want empty — a face-down spell has no text (CR 708.2a)", key)
	}
	if HasKeyword(&c, "flying") {
		t.Error("the face-down spell kept the card's printed keyword")
	}
	// CR 708.5: the controller may look; nobody else may. The stack
	// is a public ZONE and this is not a public OBJECT.
	if !c.IsKnownTo(me.ID) {
		t.Error("the caster cannot look at their own morph")
	}
	if c.IsKnownTo(opp.ID) {
		t.Error("an opponent can look at a face-down spell on the stack")
	}
	var item *StackItem
	g.ReadSnapshot(func() { item = g.StackMeta[id] })
	if item == nil || item.FaceDown != FaceDownMorphed {
		t.Fatalf("StackItem.FaceDown = %v, want %q", item, FaceDownMorphed)
	}
	if item.AltCost != "morph" {
		t.Errorf("StackItem.AltCost = %q, want morph", item.AltCost)
	}
}

// TestCastFaceDownPaysTheKeywordsThreeAndNotThePrintedCost: the {3}
// is the alternative cost (CR 702.37b) and the printed {1}{U} is not
// charged.
func TestCastFaceDownPaysTheKeywordsThreeAndNotThePrintedCost(t *testing.T) {
	g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
	var before int
	g.ReadSnapshot(func() { before = len(me.ManaPool) })
	if err := castFaceDown(t, g, me, id, "morph"); err != nil {
		t.Fatalf("cast face down: %v", err)
	}
	var after int
	g.ReadSnapshot(func() { after = len(me.ManaPool) })
	if spent := before - after; spent != 3 {
		t.Errorf("mana spent = %d, want 3 (CR 702.37b)", spent)
	}
}

// TestCastFaceDownIsACreatureSpellAtSorceryTiming is CR 708.4: it is
// a creature spell, so it needs sorcery timing even when the card
// underneath is an instant or has flash. The card here is a creature
// either way; what is asserted is that the gate saw a CREATURE
// rather than whatever the card prints.
func TestCastFaceDownIsRefusedOffTurn(t *testing.T) {
	g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
	g.WithWriteLock(func() {
		// Seat 1's turn: seat 0 has no sorcery-speed window.
		g.Turn.ActiveSeat = 1
		g.Turn.Step = StepPrecombatMain
		g.Turn.PriorityHolder = 0
	})
	err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "hand", AlternativeCost: "morph", Strict: true})
	if !errors.Is(err, ErrSorcerySpeedRequired) {
		t.Fatalf("cast face down off-turn: got %v, want ErrSorcerySpeedRequired (CR 708.4)", err)
	}
}

// TestCastFaceDownRefusesTargetsItCannotHave: a face-down spell has
// no text, so the announce gates read the CR 708.2 object rather than
// the card. A card whose real self demands a target still casts face
// down with none.
func TestCastFaceDownIgnoresThePrintedTargetClause(t *testing.T) {
	g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
	prev := CatalogTargetMode
	CatalogTargetMode = func(oracleID string) string {
		if oracleID == morphOracle {
			return "creature"
		}
		return ""
	}
	t.Cleanup(func() { CatalogTargetMode = prev })
	// Face UP the card demands a target and a cast with none is
	// refused; face down there is no clause at all.
	if err := castFaceDown(t, g, me, id, "morph"); err != nil {
		t.Fatalf("cast face down with no targets: %v", err)
	}
}

// --- 2. the resolution ------------------------------------------------

// TestAFaceDownSpellResolvesIntoAFaceDownPermanent is ADR 0082
// decision 3: the state rides the stack item onto the entry event and
// the one entry finisher applies it. The permanent is a 2/2 its
// controller alone may look at, and its ETB hook never ran.
func TestAFaceDownSpellResolvesIntoAFaceDownPermanent(t *testing.T) {
	g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
	opp := g.Seats[1]
	etbFired := false
	prev := ETBEffectHook
	ETBEffectHook = func(g *Game, cardID uuid.UUID, key string) error {
		if key == morphOracle {
			etbFired = true
		}
		return nil
	}
	t.Cleanup(func() { ETBEffectHook = prev })

	if err := castFaceDown(t, g, me, id, "morph"); err != nil {
		t.Fatalf("cast face down: %v", err)
	}
	resolveTop(t, g)

	c, zone, ok := cardAnywhere(g, id)
	if !ok || zone != ZoneBattlefield {
		t.Fatalf("card is in %q, want the battlefield", zone)
	}
	if !c.FaceDownIsPermanent() || c.FaceDownKind != FaceDownMorphed {
		t.Fatalf("face-down state = (%v, %q), want a morphed permanent", c.FaceDown, c.FaceDownKind)
	}
	if c.Controller != me.ID {
		t.Errorf("controller = %s, want the caster", c.Controller)
	}
	if !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
		t.Error("CR 708.5: the controller alone may look at a face-down permanent")
	}
	if etbFired {
		t.Error("a face-down permanent ran its card's ETB hook — CR 708.2a leaves it no text")
	}
}

// --- 3. the special action -------------------------------------------

// TestTurnFaceUpIsOfferedToTheControllerAlone is CR 708.6: the
// CONTROLLER turns a face-down permanent face up. The battlefield
// holds everybody's cards, so this is the one special action whose
// "is this yours" check is control rather than a hand lookup.
func TestTurnFaceUpIsOfferedToTheControllerAlone(t *testing.T) {
	g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
	opp := g.Seats[1]
	if err := castFaceDown(t, g, me, id, "morph"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)
	fillPool(g, opp, 4)
	err := g.PerformSpecialAction(opp.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true})
	if !errors.Is(err, ErrCardNotFound) {
		t.Fatalf("an opponent turned a morph face up: got %v, want ErrCardNotFound", err)
	}
	c, _, _ := cardAnywhere(g, id)
	if !c.FaceDown {
		t.Error("the permanent was turned face up by someone who does not control it")
	}
}

// TestTurnFaceUpIsLegalUnderSplitSecond: CR 702.61b stops players
// CASTING spells and ACTIVATING abilities, and a special action is
// neither (ADR 0082 decision 6). The asymmetry the timing table
// exists for.
func TestTurnFaceUpIsLegalUnderSplitSecond(t *testing.T) {
	g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
	if err := castFaceDown(t, g, me, id, "morph"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)
	g.WithWriteLock(func() { g.SplitSecondActive = true })
	c, _, _ := cardAnywhere(g, id)
	var ok bool
	g.ReadSnapshot(func() { ok = g.SpecialActionTimingOKLocked(me.ID, c, SpecialActionTurnFaceUp) })
	if !ok {
		t.Error("turning face up is barred under split second; CR 702.61b names casts and activations only")
	}
}

// TestTurnFaceUpIsNotOfferedOnAFaceUpPermanent: the offer is derived
// from the face-down kind, so an ordinary permanent has none and the
// verb refuses before charging.
func TestTurnFaceUpIsNotOfferedOnAFaceUpPermanent(t *testing.T) {
	g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		c := NewCard("Grizzly Bears", me.ID)
		c.InstanceID = id
		c.OracleID = morphOracle
		c.TypeLine = "Creature — Bear"
		c.Controller = me.ID
		g.Battlefield.PushTop(c)
	})
	var before int
	g.ReadSnapshot(func() { before = len(me.ManaPool) })
	err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true})
	if !errors.Is(err, ErrSpecialActionNotOffered) {
		t.Fatalf("turn a face-up permanent face up: got %v, want ErrSpecialActionNotOffered", err)
	}
	var after int
	g.ReadSnapshot(func() { after = len(me.ManaPool) })
	if before != after {
		t.Error("a refused turn-face-up charged for itself")
	}
}

// --- 4. turning face up ----------------------------------------------

// TestTurningFaceUpRestoresTheCardWithoutANewObject is CR 708.6 and
// CR 708.8 together: the real name, types and abilities come back, the
// catalog answers again, and the object is the SAME one — same
// instance, same epoch, same counters, same entry timestamp.
func TestTurningFaceUpRestoresTheCardWithoutANewObject(t *testing.T) {
	g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
	opp := g.Seats[1]
	if err := castFaceDown(t, g, me, id, "morph"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)
	before, _, _ := cardAnywhere(g, id)
	epoch, entered := before.ObjectEpoch, before.EnteredBattlefieldAt

	if err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}

	c, zone, ok := cardAnywhere(g, id)
	if !ok || zone != ZoneBattlefield {
		t.Fatalf("card is in %q, want the battlefield", zone)
	}
	if c.FaceDown || c.FaceDownKind != FaceDownNone {
		t.Fatalf("still face down: (%v, %q)", c.FaceDown, c.FaceDownKind)
	}
	if c.Name != "Willbender" {
		t.Errorf("name = %q, want the real card back", c.Name)
	}
	if key := CatalogKey(c); key != morphOracle {
		t.Errorf("CatalogKey = %q, want %q — the catalog answers again", key, morphOracle)
	}
	if !HasKeyword(&c, "flying") {
		t.Error("the card's printed keyword did not come back")
	}
	// CR 708.8: NOT a new object.
	if c.InstanceID != id {
		t.Errorf("instance id changed: %s -> %s", id, c.InstanceID)
	}
	if c.ObjectEpoch != epoch {
		t.Errorf("object epoch = %d, want %d — turning face up is not a zone change (CR 708.8)", c.ObjectEpoch, epoch)
	}
	if c.EnteredBattlefieldAt != entered {
		t.Error("the entry timestamp moved; a permanent turned face up did not just enter")
	}
	// The battlefield is public again.
	if !c.IsKnownTo(opp.ID) {
		t.Error("an opponent still cannot see a face-UP permanent")
	}
}

// TestTurningFaceUpFiresTheTurnedFaceUpTrigger is CR 708.8, and the
// ORDER that makes it possible: the state is cleared before the event
// is emitted, so the harvester reads a card with text on it.
func TestTurningFaceUpFiresTheTurnedFaceUpTrigger(t *testing.T) {
	g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
	withCatalogTriggers(t, func(oracleID string) []TriggeredAbility {
		if oracleID != morphOracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventTurnedFaceUp},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.CardID == source.InstanceID
			},
			Key: "turned face up",
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return NewTriggeredItem(source, "turned face up")
			},
		}}
	})
	if err := castFaceDown(t, g, me, id, "morph"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)
	var pending int
	g.ReadSnapshot(func() { pending = len(g.PendingTriggers) })
	if pending != 0 {
		t.Fatalf("a face-down permanent queued %d triggers on entry; CR 708.2a leaves it none", pending)
	}
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	found := false
	g.ReadSnapshot(func() {
		for _, item := range g.StackMeta {
			if item.Label == "turned face up" {
				found = true
			}
		}
		for _, item := range g.PendingTriggers {
			if item.Label == "turned face up" {
				found = true
			}
		}
	})
	if !found {
		t.Error("the CR 708.8 turned-face-up trigger did not fire")
	}
}

// --- 5. megamorph -----------------------------------------------------

// TestMegamorphPutsACounterOnAsItTurnsUp is CR 702.37b.
func TestMegamorphPutsACounterOnAsItTurnsUp(t *testing.T) {
	g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, true))
	if err := castFaceDown(t, g, me, id, "morph"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)
	down, _, _ := cardAnywhere(g, id)
	if n := down.Counters["+1/+1"]; n != 0 {
		t.Fatalf("a face-down megamorph already has %d counters", n)
	}
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	up, _, _ := cardAnywhere(g, id)
	if n := up.Counters["+1/+1"]; n != 1 {
		t.Errorf("+1/+1 counters after megamorph = %d, want 1 (CR 702.37b)", n)
	}
}

// --- 6. no way up -----------------------------------------------------

// TestTurnFaceUpOfferPerKind is ADR 0082 decision 5's table, and the
// three ways the answer is "no".
func TestTurnFaceUpOfferPerKind(t *testing.T) {
	withCatalogAlternativeCosts(t, altCostFor(morphOracle, morphOffer(FaceDownMorphed, morphCost, false)))
	creature := func(kind FaceDownKind) Card {
		c := NewCard("Willbender", uuid.New())
		c.OracleID = morphOracle
		c.TypeLine = "Creature — Human Wizard"
		c.ManaCost = "{1}{U}"
		c.SetFaceDown(kind)
		return c
	}
	land := func(kind FaceDownKind) Card {
		c := NewCard("Mountain", uuid.New())
		c.OracleID = "oracle-mountain"
		c.TypeLine = "Basic Land — Mountain"
		c.SetFaceDown(kind)
		return c
	}
	for _, tc := range []struct {
		name string
		card Card
		want string // "" means no offer at all
	}{
		{"morphed, the card prints morph: its morph cost (CR 702.37b)", creature(FaceDownMorphed), morphCost},
		{"disguised, the card prints only morph: no way up (CR 708.7)", creature(FaceDownDisguised), ""},
		{"manifested creature card: its mana cost (CR 701.40b)", creature(FaceDownManifested), "{1}{U}"},
		{"cloaked creature card: its mana cost (CR 701.58b)", creature(FaceDownCloaked), "{1}{U}"},
		{"manifested LAND: face down forever (CR 701.40b)", land(FaceDownManifested), ""},
		{"a face-UP permanent offers nothing", creature(FaceDownNone), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			offer := TurnFaceUpOffer(tc.card)
			if tc.want == "" {
				if offer != nil {
					t.Fatalf("offer = %+v, want none", offer)
				}
				return
			}
			if offer == nil {
				t.Fatal("no offer, want one")
			}
			if offer.Kind != SpecialActionTurnFaceUp {
				t.Errorf("kind = %q, want %q", offer.Kind, SpecialActionTurnFaceUp)
			}
			if offer.Cost != tc.want {
				t.Errorf("cost = %q, want %q", offer.Cost, tc.want)
			}
			// The label is the menu row and the log line, and it
			// carries the COST and not the card's name: the row is
			// shown to the controller, who can already see the card,
			// and the log line is read by a table that may not
			// (CR 708.5).
			if want := "Turn face up " + tc.want; offer.Label != want {
				t.Errorf("label = %q, want %q", offer.Label, want)
			}
		})
	}
}

// TestAFaceDownPermanentIsRevealedWhenItLeaves is CR 708.9, asserted
// here for a MORPH rather than for the manifest ADR 0069 tested: the
// reveal is on the route, so it covers every way a face-down
// permanent can leave.
func TestAMorphIsRevealedWhenItLeavesTheBattlefield(t *testing.T) {
	g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
	opp := g.Seats[1]
	if err := castFaceDown(t, g, me, id, "morph"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)
	g.WithWriteLock(func() {
		if _, err := g.routeCardToZoneLocked(zoneRoute{CardID: id, Dst: ZoneGraveyard, Actor: me.ID}); err != nil {
			t.Fatalf("route: %v", err)
		}
	})
	c, zone, ok := cardAnywhere(g, id)
	if !ok || zone != ZoneGraveyard {
		t.Fatalf("card is in %q, want the graveyard", zone)
	}
	if c.FaceDown {
		t.Error("the card is still face down in the graveyard; MoveCard clears it (CR 400.7)")
	}
	if !c.IsKnownTo(opp.ID) {
		t.Error("CR 708.9: a face-down permanent is revealed as it leaves the battlefield")
	}
}

// TestACounteredMorphIsRevealedAsItLeavesTheStack is the exit CR 708.9
// covers that had no way to happen before the face-down CAST existed:
// a face-down SPELL leaving the stack. The table watched a {3} it
// could not read, and it finds out what the {3} was buying.
//
// The log reason has to name the zone the object actually left:
// telling the table a countered spell "left the battlefield" would be
// describing a game that did not happen.
func TestACounteredMorphIsRevealedAsItLeavesTheStack(t *testing.T) {
	g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, false))
	opp := g.Seats[1]
	if err := castFaceDown(t, g, me, id, "morph"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := g.CounterSpell(id, nil); err != nil {
		t.Fatalf("counter: %v", err)
	}
	c, zone, ok := cardAnywhere(g, id)
	if !ok || zone != ZoneGraveyard {
		t.Fatalf("card is in %q, want the graveyard", zone)
	}
	if c.FaceDown {
		t.Error("the countered spell is still face down; MoveCard clears it (CR 400.7)")
	}
	if !c.IsKnownTo(opp.ID) {
		t.Error("CR 708.9: a face-down spell is revealed as it leaves the stack")
	}
	const want = "turned face up as it left the stack"
	found := false
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventRevealCards && ev.CardID == id {
				found = true
				if ev.Label != want {
					t.Errorf("reveal reason = %q, want %q", ev.Label, want)
				}
			}
		}
	})
	if !found {
		t.Error("no reveal event for the countered morph")
	}
}

package game

import (
	"testing"

	"github.com/google/uuid"
)

// face_down_turn_test.go — #1209, ADR 0082's 2026-09-23 amendment:
// turning a permanent that is already on the battlefield FACE DOWN
// (CR 708.2a).
//
// The assertions that matter are the ones about what this does NOT
// do. Turning the state on is two lines and is hard to get wrong; the
// design risk is that somebody later implements it as an exile-and-
// return or as a re-entry, and every test in section 2 is written so
// that either would fail it.
//
//	1. the CR 708.2a body, and the kind that records why
//	2. not a new object — counters, damage, combat, the timestamp
//	3. the refusals: CR 708.2b and CR 712.16
//	4. who may look (CR 708.5), and the batch's one moment
//	5. the way back up, or the lack of one (CR 708.7 / CR 702.37e)
//	6. tokens, the CR 708.9 reveal, undo and the snapshot

// turnableCreature seats a face-up, PUBLIC creature on the
// battlefield the way a resolved permanent spell would: stamped with
// its CR 613.7 timestamp and known to every seat, because the
// battlefield is a public zone and that is the state the turn-face-
// down has to take away.
func turnableCreature(t *testing.T, g *Game, owner *Player, name string) uuid.UUID {
	t.Helper()
	c := NewCard(name, owner.ID)
	c.Controller = owner.ID
	c.TypeLine = "Creature — Demon"
	c.OracleID = "oracle-" + name
	c.ManaCost = "{4}{B}{B}"
	c.Colors = []string{"B"}
	c.Power, c.Toughness = 7, 7
	c.Keywords = []string{"flying", "deathtouch"}
	id := c.InstanceID
	g.Battlefield.PushTop(c)
	stampBattlefieldEntryLocked(g, id)
	g.markCardKnownInZoneLocked(g.Battlefield, id)
	return id
}

// --- 1. the body and the kind ----------------------------------------

// TestTurnFaceDownMakesTheCR7082aBody is the headline. CR 708.2a: "If
// a face-up permanent is turned face down by a spell or ability that
// doesn't list any characteristics for that object, it becomes a 2/2
// face-down creature with no text, no name, no subtypes, and no mana
// cost."
//
// Nothing here is new code — the body, the catalog silence and the
// four accessors are ADR 0069's — which is the point: the primitive
// sets one field and the whole projection follows.
func TestTurnFaceDownMakesTheCR7082aBody(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := turnableCreature(t, g, me, "Ixidron")
	id := turnableCreature(t, g, me, "Sheoldred")

	turned := g.TurnFaceDownForEffect(src, id)
	if len(turned) != 1 || turned[0] != id {
		t.Fatalf("TurnFaceDownForEffect returned %v, want [%s]", turned, id)
	}

	c := findBattlefieldCard(g, id)
	if c == nil {
		t.Fatal("the permanent left the battlefield — turning face down is not a zone change")
	}
	if !c.FaceDown || c.FaceDownKind != FaceDownTurned {
		t.Fatalf("face-down state = (%v, %q), want (true, %q)", c.FaceDown, c.FaceDownKind, FaceDownTurned)
	}
	if !c.FaceDownIsPermanent() {
		t.Error("the new kind is not a CR 708.2 permanent state — the 2/2 projection will not reach it")
	}
	if got := CatalogKey(*c); got != "" {
		t.Errorf("CatalogKey = %q, want \"\" — CR 708.2a leaves it with no text", got)
	}
	if got := c.printedCharacteristic(); got.Power != 2 || got.Toughness != 2 ||
		got.Name != "" || len(got.Types) != 1 || got.Types[0] != "Creature" || len(got.Colors) != 0 {
		t.Errorf("CR 708.2a body = %+v, want a nameless colourless 2/2 Creature", got)
	}
	if HasKeyword(c, "flying") {
		t.Error("the face-down permanent kept flying — CR 708.2a is \"no text\"")
	}
	if c.FaceDownKind.HasWard() {
		t.Error("a permanent turned face down has ward — CR 708.2a lists no characteristics, " +
			"and ward {2} is disguise's and cloak's (CR 702.168a, CR 701.58a)")
	}
}

// --- 2. not a new object ---------------------------------------------

// TestTurnFaceDownIsNotANewObject. CR 613.7f gives a permanent a new
// timestamp "each time it turns face up or face down" — a statement
// about ONE permanent — and CR 708.8 says the same thing for the other
// direction in as many words. Every field checked here is one MoveCard
// strips on a battlefield exit, so an implementation that reached for
// exile-and-return would fail all of them at once.
func TestTurnFaceDownIsNotANewObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := turnableCreature(t, g, me, "Sheoldred")

	card := findBattlefieldCard(g, id)
	card.Tapped = true
	card.Counters = map[string]int{"+1/+1": 3}
	card.DamageMarked = 2
	card.AttackingTarget = g.Seats[1].ID
	before := *card

	g.TurnFaceDownForEffect(uuid.Nil, id)

	after := findBattlefieldCard(g, id)
	if after == nil {
		t.Fatal("the permanent left the battlefield")
	}
	if after.InstanceID != id {
		t.Error("InstanceID changed — the permanent did not become a new object")
	}
	if after.ObjectEpoch != before.ObjectEpoch {
		t.Errorf("ObjectEpoch = %d, want %d", after.ObjectEpoch, before.ObjectEpoch)
	}
	if !after.Tapped {
		t.Error("the permanent untapped")
	}
	if after.Counters["+1/+1"] != 3 {
		t.Errorf("+1/+1 counters = %d, want 3 — counters ride through", after.Counters["+1/+1"])
	}
	if after.DamageMarked != 2 {
		t.Errorf("marked damage = %d, want 2", after.DamageMarked)
	}
	if after.AttackingTarget != g.Seats[1].ID {
		t.Error("the permanent was removed from combat — an Ixidron'd attacker goes on attacking")
	}
	if after.EnteredBattlefieldAt != before.EnteredBattlefieldAt {
		t.Error("the CR 613.7 entry timestamp was re-stamped — the permanent did not re-enter")
	}
	if after.SummonedThisTurn != before.SummonedThisTurn {
		t.Error("summoning sickness changed")
	}
	if after.Controller != me.ID {
		t.Error("the controller changed — CR 708.2a does not touch layer 2")
	}
	for _, ev := range g.Events {
		switch ev.Kind {
		case EventZoneMove, EventETB, EventLTB:
			if ev.CardID == id {
				t.Errorf("turning face down emitted %s — nothing entered or left", ev.Kind)
			}
		}
	}
}

// TestTurnFaceDownEmitsItsEventAndBumpsTheLayerVersion pins the two
// engine-side consequences everything downstream depends on: the
// layer engine has to be told that the printed characteristics were
// replaced wholesale, and a trigger on another permanent has to have
// something to watch.
func TestTurnFaceDownEmitsItsEventAndBumpsTheLayerVersion(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := turnableCreature(t, g, me, "Ixidron")
	id := turnableCreature(t, g, me, "Sheoldred")
	g.RecomputeLayersIfStaleLocked()
	if got := findBattlefieldCard(g, id).Effective().Power; got != 7 {
		t.Fatalf("fixture power = %d, want 7 — the cache was never warm, so this proves nothing", got)
	}
	before := g.layerVersion.Load()

	g.TurnFaceDownForEffect(src, id)

	if g.layerVersion.Load() <= before {
		t.Error("layerVersion did not move — the CR 708.2a body replaced the printed one")
	}
	g.RecomputeLayersIfStaleLocked()
	if got := findBattlefieldCard(g, id).Effective().Power; got != 2 {
		t.Errorf("effective power = %d, want 2 — the layer cache did not see the state change", got)
	}
	var found *Event
	for i := range g.Events {
		if g.Events[i].Kind == EventTurnedFaceDown && g.Events[i].CardID == id {
			found = &g.Events[i]
		}
	}
	if found == nil {
		t.Fatal("no EventTurnedFaceDown was emitted")
	}
	if found.Actor != me.ID {
		t.Errorf("Actor = %v, want the permanent's controller %v", found.Actor, me.ID)
	}
	if found.Source != src {
		t.Errorf("Source = %v, want the object that did it %v", found.Source, src)
	}
}

// --- 3. the refusals --------------------------------------------------

// TestTurnFaceDownRefusals — both are "nothing happens" in the rules,
// so both are a skip rather than an error, and neither may emit an
// event that a trigger could fire off.
func TestTurnFaceDownRefusals(t *testing.T) {
	t.Run("CR 708.2b: a face-down permanent can't be turned face down", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		id := turnableCreature(t, g, me, "Sheoldred")
		// Put it face down as a MORPH first, so the assertion is that
		// the second turn leaves the kind (and therefore the way back
		// up) alone rather than overwriting it with `turned`.
		g.applyFaceDownLandingLocked(g.Battlefield, id, FaceDownMorphed)

		if got := g.TurnFaceDownForEffect(uuid.Nil, id); len(got) != 0 {
			t.Errorf("turned %v, want nothing — CR 708.2b", got)
		}
		if k := findBattlefieldCard(g, id).FaceDownKind; k != FaceDownMorphed {
			t.Errorf("kind = %q, want it unchanged at %q — CR 708.2b says the effect "+
				"\"doesn't change any of its characteristics or their copiable values\"", k, FaceDownMorphed)
		}
		if hasEvent(g, EventTurnedFaceDown, id) {
			t.Error("a refused turn emitted EventTurnedFaceDown — nothing happened")
		}
	})

	t.Run("CR 712.16: a double-faced permanent can't be turned face down", func(t *testing.T) {
		g := newActiveGame(t)
		me := g.Seats[0]
		id := pushTransformFixture(t, g, transformCreatureFixture(me.ID))

		if got := g.TurnFaceDownForEffect(uuid.Nil, id); len(got) != 0 {
			t.Errorf("turned %v, want nothing — CR 712.16", got)
		}
		c := findBattlefieldCard(g, id)
		if c.FaceDown {
			t.Error("a double-faced permanent was turned face down")
		}
		if c.Name != "Fixture Villager" {
			t.Errorf("name = %q — the refusal must not change any characteristic", c.Name)
		}
	})

	t.Run("a card that is not on the battlefield is skipped", func(t *testing.T) {
		g := newActiveGame(t)
		if got := g.TurnFaceDownForEffect(uuid.Nil, uuid.New()); len(got) != 0 {
			t.Errorf("turned %v, want nothing", got)
		}
	})
}

// --- 4. who may look, and the batch -----------------------------------

// TestTurnFaceDownResetsTheKnowersToTheController is CR 708.5: "you
// can't look at face-down cards in any other zone or face-down spells
// or permanents controlled by another player."
//
// The permanent was PUBLIC a moment ago — every seat could read it —
// so this is a genuine narrowing, and it is the one place the engine
// deliberately forgets something a paper table would remember. ADR
// 0082's amendment decision A3 is the argument.
func TestTurnFaceDownResetsTheKnowersToTheController(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := turnableCreature(t, g, me, "Sheoldred")
	if !findBattlefieldCard(g, id).IsKnownTo(opp.ID) {
		t.Fatal("fixture: the permanent was not public to begin with, so the narrowing proves nothing")
	}

	g.TurnFaceDownForEffect(uuid.Nil, id)

	c := findBattlefieldCard(g, id)
	if !c.IsKnownTo(me.ID) {
		t.Error("the controller may not look at their own face-down permanent (CR 708.5)")
	}
	if c.IsKnownTo(opp.ID) {
		t.Error("an opponent may still look at a permanent that has been turned face down (CR 708.5)")
	}
}

// batchProbe fails the moment it is told about one permanent turning
// face down while another in the same batch has not turned yet.
type batchProbe struct {
	watch   []uuid.UUID
	g       *Game
	partial bool
	seen    int
}

func (p *batchProbe) OnEvent(g *Game, ev Event) {
	if ev.Kind != EventTurnedFaceDown {
		return
	}
	p.seen++
	for _, id := range p.watch {
		c := findBattlefieldCard(g, id)
		if c == nil || !c.FaceDown {
			p.partial = true
		}
	}
}

// TestTurnFaceDownTurnsTheWholeBatchBeforeAnnouncingAnyOfIt is
// Ixidron's requirement and phaseOutLocked's contract: "turn all other
// nontoken creatures face down" is one event in the game, so a trigger
// watching it must never see a board that is half turned over.
func TestTurnFaceDownTurnsTheWholeBatchBeforeAnnouncingAnyOfIt(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := turnableCreature(t, g, me, "Ixidron")
	a := turnableCreature(t, g, me, "Sheoldred")
	b := turnableCreature(t, g, opp, "Griselbrand")
	probe := &batchProbe{watch: []uuid.UUID{a, b}, g: g}
	g.RegisterListener(probe)

	turned := g.TurnFaceDownForEffect(src, a, b)

	if len(turned) != 2 {
		t.Fatalf("turned %d permanents, want 2", len(turned))
	}
	if probe.seen != 2 {
		t.Errorf("saw %d events, want one per permanent", probe.seen)
	}
	if probe.partial {
		t.Error("an event went out while part of the batch was still face up")
	}
	// The event's Actor is the permanent's OWN controller, which is
	// what makes an opponent's creature read as theirs in the log.
	for _, ev := range g.Events {
		if ev.Kind == EventTurnedFaceDown && ev.CardID == b && ev.Actor != opp.ID {
			t.Errorf("Actor for the opponent's permanent = %v, want %v", ev.Actor, opp.ID)
		}
	}
}

// TestTurnFaceDownTakesEachPermanentOnce — a repeated ID is one
// permanent, and the second mention must not be read as CR 708.2b
// firing on the engine's own work either.
func TestTurnFaceDownTakesEachPermanentOnce(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := turnableCreature(t, g, me, "Sheoldred")
	if got := g.TurnFaceDownForEffect(uuid.Nil, id, id); len(got) != 1 {
		t.Errorf("turned %v, want one entry", got)
	}
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == EventTurnedFaceDown && ev.CardID == id {
			n++
		}
	}
	if n != 1 {
		t.Errorf("emitted %d events for one permanent, want 1", n)
	}
}

// --- 5. the way back up ------------------------------------------------

// TestTurnFaceUpOfferForAPermanentTurnedFaceDown is CR 708.7's two
// halves at once. The effect that turned it over gave it no way back
// up, so whether there is one at all is a question about the CARD:
// CR 702.37e ("a face-down permanent you control WITH A MORPH
// ABILITY") and CR 702.168d say the same for disguise, and both key
// on the card having the ability rather than on how the permanent got
// face down.
func TestTurnFaceUpOfferForAPermanentTurnedFaceDown(t *testing.T) {
	for _, tc := range []struct {
		name        string
		declares    *AlternativeCost
		wantOffer   bool
		wantCost    string
		wantCounter bool
	}{
		{
			name:      "CR 708.7: a card with no face-down cast can never be turned face up",
			declares:  nil,
			wantOffer: false,
		},
		{
			name:      "CR 702.37e: morph is the card's, whatever turned it over",
			declares:  ptrAltCost(morphOffer(FaceDownMorphed, morphCost, false)),
			wantOffer: true, wantCost: morphCost,
		},
		{
			name:      "CR 702.109b: megamorph still pays its counter",
			declares:  ptrAltCost(morphOffer(FaceDownMorphed, morphCost, true)),
			wantOffer: true, wantCost: morphCost, wantCounter: true,
		},
		{
			name:      "CR 702.168d: disguise is the same rule with the other keyword",
			declares:  ptrAltCost(morphOffer(FaceDownDisguised, morphCost, false)),
			wantOffer: true, wantCost: morphCost,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.declares != nil {
				withCatalogAlternativeCosts(t, altCostFor(morphOracle, *tc.declares))
			} else {
				withCatalogAlternativeCosts(t, altCostFor(morphOracle))
			}
			g := newActiveGame(t)
			me := g.Seats[0]
			id := turnableCreature(t, g, me, "Sheoldred")
			findBattlefieldCard(g, id).OracleID = morphOracle
			g.TurnFaceDownForEffect(uuid.Nil, id)

			got := TurnFaceUpOffer(*findBattlefieldCard(g, id))
			if tc.wantOffer != (got != nil) {
				t.Fatalf("TurnFaceUpOffer = %v, want offered = %v", got, tc.wantOffer)
			}
			if got == nil {
				return
			}
			if got.Cost != tc.wantCost {
				t.Errorf("cost = %q, want the card's %q", got.Cost, tc.wantCost)
			}
			if got.FaceUpCounter != tc.wantCounter {
				t.Errorf("FaceUpCounter = %v, want %v", got.FaceUpCounter, tc.wantCounter)
			}
		})
	}
}

// ptrAltCost is a one-liner so the table above can say "declares
// nothing" with a nil rather than with a second bool.
func ptrAltCost(a AlternativeCost) *AlternativeCost { return &a }

// TestATurnedMorphCanActuallyBeTurnedFaceUp walks the whole Backslide
// loop through the public special action, because the offer being
// right is only half of it: the performer re-checks the state for
// itself, and the card has to come back whole.
func TestATurnedMorphCanActuallyBeTurnedFaceUp(t *testing.T) {
	withCatalogAlternativeCosts(t, altCostFor(morphOracle, morphOffer(FaceDownMorphed, morphCost, false)))
	g := newActiveGame(t)
	me := g.Seats[0]
	id := turnableCreature(t, g, me, "Willbender")
	findBattlefieldCard(g, id).OracleID = morphOracle
	g.TurnFaceDownForEffect(uuid.Nil, id)
	fillPool(g, me, 8)
	g.WithWriteLock(func() { g.Turn.Step = StepPrecombatMain })

	if err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	c := findBattlefieldCard(g, id)
	if c.FaceDown {
		t.Fatal("the permanent is still face down")
	}
	if c.Name != "Willbender" || !HasKeyword(c, "flying") {
		t.Errorf("the real card did not come back: %q, flying = %v", c.Name, HasKeyword(c, "flying"))
	}
	if !hasEvent(g, EventTurnedFaceUp, id) {
		t.Error("no EventTurnedFaceUp — a \"when this is turned face up\" trigger has nothing to watch")
	}
}

// --- 6. tokens, the reveal, and undo ----------------------------------

// TestAFaceDownTokenIsStillAToken. Nothing in CR 708 or CR 111 stops a
// token being turned face down, and the trap is that the CR 708.2a
// body has no Token supertype in it: if IsToken read the projection
// instead of the printed line, CR 704.5d would stop applying and a
// bounced face-down token would hand its controller a real card.
func TestAFaceDownTokenIsStillAToken(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushIntrinsicPermanent(g, me, "Zombie", "Token Creature — Zombie", nil, nil)
	g.markCardKnownInZoneLocked(g.Battlefield, id)

	if got := g.TurnFaceDownForEffect(uuid.Nil, id); len(got) != 1 {
		t.Fatalf("turned %v, want the token — no rule refuses one", got)
	}
	c := findBattlefieldCard(g, id)
	if !c.IsToken() {
		t.Fatal("a face-down token stopped reading as a token — CR 704.5d would leak it into a hand")
	}
	if !c.FaceDownIsPermanent() || c.printedCharacteristic().Power != 2 {
		t.Error("the token did not get the CR 708.2a body")
	}
}

// TestTurnFaceDownStillRevealsOnLeavingTheBattlefield is CR 708.9,
// which ADR 0069 already built: "if a face-down permanent moves from
// the battlefield to any other zone, its owner must reveal it to all
// players as they move it." The new kind is a CR 708.2 permanent
// state, so it gets the reveal with no code of its own — and that is
// exactly the assertion, because a seventh kind that forgot to join
// IsPermanentState would silently lose it.
func TestTurnFaceDownStillRevealsOnLeavingTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := turnableCreature(t, g, me, "Sheoldred")
	g.TurnFaceDownForEffect(uuid.Nil, id)
	if findBattlefieldCard(g, id).IsKnownTo(opp.ID) {
		t.Fatal("fixture: the opponent already knows it, so the reveal proves nothing")
	}

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})

	if !hasEvent(g, EventRevealCards, id) {
		t.Error("no CR 708.9 reveal as the face-down permanent left the battlefield")
	}
}

// TestUndoRestoresAPermanentTurnedFaceUpAgain: an undo restores a
// clone, so the new kind has to ride it. Without the field on the
// clone, rewinding an Ixidron would leave the board face down for
// ever.
func TestUndoRestoresAPermanentTurnedFaceUpAgain(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := turnableCreature(t, g, me, "Sheoldred")
	before := g.Clone()

	g.TurnFaceDownForEffect(uuid.Nil, id)
	if !findBattlefieldCard(g, id).FaceDown {
		t.Fatal("fixture: nothing turned face down, so the undo proves nothing")
	}

	g.RestoreFrom(before)

	c := findBattlefieldCard(g, id)
	if c == nil {
		t.Fatal("the permanent is gone after the undo")
	}
	if c.FaceDown || c.FaceDownKind != FaceDownNone {
		t.Errorf("after undo the state is (%v, %q), want face up", c.FaceDown, c.FaceDownKind)
	}
	if c.Name != "Sheoldred" || c.CurrentPower() != 7 {
		t.Errorf("after undo the object is %q %d/%d, want the real card back",
			c.Name, c.CurrentPower(), c.CurrentToughness())
	}
	if !c.IsKnownTo(opp.ID) {
		t.Error("after undo the opponent still may not look at a public permanent — " +
			"the knower set did not rewind with the state")
	}
}

// TestSnapshotCarriesAPermanentTurnedFaceDown is the persistence half
// of the same property: a saved game reloads with the permanent still
// face down and still for the same reason, which is what
// TurnFaceUpOffer reads to decide whether there is a way back up.
func TestSnapshotCarriesAPermanentTurnedFaceDown(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[0]
	id := turnableCreature(t, g, me, "Sheoldred")
	g.TurnFaceDownForEffect(uuid.Nil, id)

	_, restored := roundTrip(t, g)

	c := findBattlefieldCard(restored, id)
	if c == nil {
		t.Fatal("the permanent did not survive the snapshot")
	}
	if !c.FaceDown || c.FaceDownKind != FaceDownTurned {
		t.Errorf("restored state = (%v, %q), want (true, %q)", c.FaceDown, c.FaceDownKind, FaceDownTurned)
	}
	if !c.FaceDownIsPermanent() || c.printedCharacteristic().Power != 2 {
		t.Error("the restored permanent lost its CR 708.2a body")
	}
}

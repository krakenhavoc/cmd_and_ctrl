package game

import (
	"testing"

	"github.com/google/uuid"
)

// face_down_turn_up_effect_test.go — #2590, ADR 0082's second
// 2026-10-07 amendment: turning a face-down permanent face up as part
// of an EFFECT (CR 708.8, CR 701.40b), for free.
//
// The special action's own tests live in face_down_turn_test.go and
// special_action_test.go. What is tested here is the door that skips
// the price, and the two places it must still agree with the action:
// the object it leaves behind and the event it emits.

// faceDownPermanent seats a face-down permanent of the given kind whose
// card is `typeLine`, controlled and known only by `owner`.
func faceDownPermanent(t *testing.T, g *Game, owner *Player, name, typeLine string, kind FaceDownKind) uuid.UUID {
	t.Helper()
	c := NewCard(name, owner.ID)
	c.Controller = owner.ID
	c.TypeLine = typeLine
	c.OracleID = "oracle-" + name
	c.ManaCost = "{2}{G}"
	if typeLine != "Instant" && typeLine != "Land" {
		c.Power, c.Toughness = 3, 3
	}
	id := c.InstanceID
	g.Battlefield.PushTop(c)
	stampBattlefieldEntryLocked(g, id)
	g.applyFaceDownLandingLocked(g.Battlefield, id, kind, nil)
	return id
}

// The headline: an effect turns a face-down permanent face up, pays
// nothing, and the permanent is the same object, public again.
func TestTurnFaceUpForEffectPaysNothingAndIsNotANewObject(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := turnableCreature(t, g, me, "Shrieker")
	id := faceDownPermanent(t, g, opp, "Sleeper", "Creature — Horror", FaceDownManifested)
	card := findBattlefieldCard(g, id)
	card.Counters = map[string]int{"+1/+1": 2}
	card.DamageMarked = 1
	card.Tapped = true
	before := *card
	fillPool(g, me, 5)
	manaBefore := len(me.ManaPool)

	if !g.TurnFaceUpForEffect(src, me.ID, id) {
		t.Fatal("TurnFaceUpForEffect refused a manifested creature card")
	}

	after := findBattlefieldCard(g, id)
	if after == nil || after.FaceDown {
		t.Fatal("the permanent is not face up on the battlefield")
	}
	if after.Name != "Sleeper" || after.Effective().Power != 3 {
		t.Errorf("the real card did not come back: %q %d", after.Name, after.Effective().Power)
	}
	if after.InstanceID != id || after.ObjectEpoch != before.ObjectEpoch ||
		after.EnteredBattlefieldAt != before.EnteredBattlefieldAt {
		t.Error("turning face up made a new object (CR 708.8)")
	}
	if after.Counters["+1/+1"] != 2 || after.DamageMarked != 1 || !after.Tapped {
		t.Error("counters, damage or tap state did not ride through")
	}
	if after.Controller != opp.ID {
		t.Errorf("controller = %v, want it unchanged (%v): the actor is not the controller", after.Controller, opp.ID)
	}
	if after.FaceTurnedAt == 0 || after.FaceTurnedAt == before.FaceTurnedAt {
		t.Error("CR 613.7f: turning face up must renew the timestamp")
	}
	for _, p := range g.Seats {
		if !after.IsKnownTo(p.ID) {
			t.Errorf("seat %s does not know a face-up permanent", p.Name)
		}
	}
	if len(me.ManaPool) != manaBefore {
		t.Error("an effect's turn paid mana")
	}
}

// The event is the special action's, with the effect's object as the
// Source and the effect's player as the Actor, so "when turned face up"
// and Growing Dread's "whenever you turn a permanent face up" fire.
func TestTurnFaceUpForEffectEmitsTheSameEventAfterTheStateIsCleared(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := turnableCreature(t, g, me, "Shrieker")
	id := faceDownPermanent(t, g, me, "Sleeper", "Creature — Horror", FaceDownManifested)
	g.RecomputeLayersIfStaleLocked()
	if got := findBattlefieldCard(g, id).Effective().Power; got != 2 {
		t.Fatalf("fixture power = %d, want the 2/2 body", got)
	}
	version := g.layerVersion.Load()

	probe := &faceUpProbe{id: id}
	g.RegisterListener(probe)

	if !g.TurnFaceUpForEffect(src, me.ID, id) {
		t.Fatal("refused")
	}

	var ev *Event
	for i := range g.Events {
		if g.Events[i].Kind == EventTurnedFaceUp && g.Events[i].CardID == id {
			ev = &g.Events[i]
		}
	}
	if ev == nil {
		t.Fatal("no EventTurnedFaceUp was emitted")
	}
	if ev.Actor != me.ID || ev.Source != src {
		t.Errorf("event actor/source = %v/%v, want %v/%v", ev.Actor, ev.Source, me.ID, src)
	}
	if probe.faceDownAtEmit {
		t.Error("the event went out while the permanent was still face down — the trigger harvest would find no text")
	}
	if g.layerVersion.Load() <= version {
		t.Error("layerVersion did not move")
	}
	g.RecomputeLayersIfStaleLocked()
	if got := findBattlefieldCard(g, id).Effective().Power; got != 3 {
		t.Errorf("effective power = %d, want the real 3", got)
	}
}

type faceUpProbe struct {
	id             uuid.UUID
	faceDownAtEmit bool
}

func (p *faceUpProbe) OnEvent(g *Game, ev Event) {
	if ev.Kind != EventTurnedFaceUp || ev.CardID != p.id {
		return
	}
	if c := findBattlefieldCard(g, p.id); c != nil && c.FaceDown {
		p.faceDownAtEmit = true
	}
}

// CR 701.40b: a manifested card that is not a creature card stays
// face down. An effect that tries does nothing: no change, no event.
func TestTurnFaceUpForEffectRefusesAManifestedNoncreatureCard(t *testing.T) {
	for _, kind := range []FaceDownKind{FaceDownManifested, FaceDownCloaked} {
		for _, typeLine := range []string{"Land", "Instant"} {
			g := newActiveGame(t)
			me := g.Seats[0]
			id := faceDownPermanent(t, g, me, "Dud", typeLine, kind)

			if CanTurnFaceUpForEffect(*findBattlefieldCard(g, id)) {
				t.Errorf("%s %s: CanTurnFaceUpForEffect says yes", kind, typeLine)
			}
			if g.TurnFaceUpForEffect(me.ID, me.ID, id) {
				t.Errorf("%s %s: turned face up", kind, typeLine)
			}
			c := findBattlefieldCard(g, id)
			if !c.FaceDown || c.FaceDownKind != kind {
				t.Errorf("%s %s: the refused permanent changed state", kind, typeLine)
			}
			if hasEvent(g, EventTurnedFaceUp, id) {
				t.Errorf("%s %s: a refused turn emitted EventTurnedFaceUp", kind, typeLine)
			}
		}
	}
}

// The refusals that are not the card's: a face-up permanent, a card
// that is gone, and uuid.Nil are all "nothing happens".
func TestTurnFaceUpForEffectIgnoresWhatIsNotFaceDown(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	up := turnableCreature(t, g, me, "Plain")
	if g.TurnFaceUpForEffect(me.ID, me.ID, up) {
		t.Error("turned a face-up permanent face up")
	}
	if g.TurnFaceUpForEffect(me.ID, me.ID, uuid.New()) || g.TurnFaceUpForEffect(me.ID, me.ID, uuid.Nil) {
		t.Error("turned a permanent that is not on the battlefield")
	}
	if hasEvent(g, EventTurnedFaceUp, up) {
		t.Error("emitted EventTurnedFaceUp for a permanent that was never face down")
	}
}

// CR 708.7 withholds the special action from a permanent an effect
// turned face down when its card prints no way back up. It does not
// withhold an EFFECT: this is the case TurnFaceUpOffer answers nil
// to and the effect door answers yes to.
func TestTurnFaceUpForEffectTurnsWhatTheSpecialActionCannot(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := turnableCreature(t, g, me, "Ixidron")
	id := turnableCreature(t, g, me, "Sheoldred")
	g.TurnFaceDownForEffect(src, id)
	if TurnFaceUpOffer(*findBattlefieldCard(g, id)) != nil {
		t.Fatal("fixture: Sheoldred should have no way back up by the special action")
	}

	if !g.TurnFaceUpForEffect(src, me.ID, id) {
		t.Fatal("the effect door refused a permanent the special action cannot reach")
	}
	if findBattlefieldCard(g, id).FaceDown {
		t.Error("still face down")
	}
}

// CR 702.37b: the megamorph counter is owed only if the megamorph COST
// was paid to turn it face up. The special action pays it; an effect
// does not, and so does not owe the counter.
func TestTurnFaceUpForEffectGivesNoMegamorphCounter(t *testing.T) {
	withCatalogAlternativeCosts(t, altCostFor(morphOracle, morphOffer(FaceDownMorphed, morphCost, true)))

	// By effect: no counter.
	g := newActiveGame(t)
	me := g.Seats[0]
	id := turnableCreature(t, g, me, "Hooded")
	findBattlefieldCard(g, id).OracleID = morphOracle
	g.applyFaceDownLandingLocked(g.Battlefield, id, FaceDownMorphed, nil)
	if !g.TurnFaceUpForEffect(me.ID, me.ID, id) {
		t.Fatal("refused")
	}
	if n := findBattlefieldCard(g, id).Counters["+1/+1"]; n != 0 {
		t.Errorf("an effect's turn put %d +1/+1 counters on a megamorph card", n)
	}

	// By the special action: the counter.
	g2 := newActiveGame(t)
	me2 := g2.Seats[0]
	id2 := turnableCreature(t, g2, me2, "Hooded")
	findBattlefieldCard(g2, id2).OracleID = morphOracle
	g2.applyFaceDownLandingLocked(g2.Battlefield, id2, FaceDownMorphed, nil)
	fillPool(g2, me2, 8)
	g2.WithWriteLock(func() { g2.Turn.Step = StepPrecombatMain })
	if err := g2.PerformSpecialAction(me2.ID, id2, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("special action: %v", err)
	}
	if n := findBattlefieldCard(g2, id2).Counters["+1/+1"]; n != 1 {
		t.Errorf("the special action put %d counters, want 1", n)
	}
}

// The special action still names the permanent as its own Source, so
// the log stays silent on it (its LogSpecialAction line carries it).
func TestTheSpecialActionEventKeepsTheCardAsItsOwnSource(t *testing.T) {
	withCatalogAlternativeCosts(t, altCostFor(morphOracle, morphOffer(FaceDownMorphed, morphCost, false)))
	g := newActiveGame(t)
	me := g.Seats[0]
	id := turnableCreature(t, g, me, "Willbender")
	findBattlefieldCard(g, id).OracleID = morphOracle
	g.applyFaceDownLandingLocked(g.Battlefield, id, FaceDownMorphed, nil)
	fillPool(g, me, 8)
	g.WithWriteLock(func() { g.Turn.Step = StepPrecombatMain })
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	for _, ev := range g.Events {
		if ev.Kind == EventTurnedFaceUp && ev.CardID == id && (ev.Source != id || ev.Actor != me.ID) {
			t.Errorf("event source/actor = %v/%v, want the card itself and its controller", ev.Source, ev.Actor)
		}
	}
}

package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_rebound_cards_test.go — the cards that GIVE a spell rebound
// (#1854, ADR 0107 §3 decision 5). The mechanic is pinned in
// game/granted_rebound_test.go; what is pinned here is that each
// catalog declaration reaches it.

const (
	castThroughTimeOracle = "0a911cb2-caae-4378-bf09-1c5b0751dd35"
	taigamOracle          = "e4cf3710-8600-4f95-abb2-faefdc25693d"
)

// reboundTriggersFor counts the queued rebound delayed triggers for
// the exiled card `id`.
func reboundTriggersFor(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.Body.Key() == "rebound/cast" && dt.Params.Object.ID == id {
			n++
		}
	}
	return n
}

// castPlainSpell casts a {0} spell of `typeLine` with no rebound of its
// own from p's hand.
func castPlainSpell(t *testing.T, g *game.Game, p *game.Player, typeLine string) uuid.UUID {
	t.Helper()
	return castFromHandForTest(t, g, p, "Plain "+typeLine, typeLine, "{0}", "", game.CastSpellParams{})
}

// Cast Through Time: an instant its controller casts from hand is
// exiled as it resolves and rebounds; a creature spell is not touched,
// and the opponent's instant goes to the graveyard.
func TestCastThroughTimeGivesYourInstantsAndSorceriesRebound(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Cast Through Time", "Enchantment", castThroughTimeOracle, false)

	instant := castPlainSpell(t, g, me, "Instant")
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(instant) || me.Graveyard.Contains(instant) {
		t.Fatal("an instant cast from hand under Cast Through Time was not exiled by rebound")
	}
	if n := reboundTriggersFor(g, instant); n != 1 {
		t.Errorf("rebound triggers = %d, want 1", n)
	}

	sorcery := castPlainSpell(t, g, me, "Sorcery")
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(sorcery) || reboundTriggersFor(g, sorcery) != 1 {
		t.Error("a sorcery cast from hand under Cast Through Time did not rebound")
	}

	g.WithWriteLock(func() { g.Turn.PriorityHolder = (seat + 1) % len(g.Seats) })
	theirs := castPlainSpell(t, g, opp, "Instant")
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's instant rebounded off your Cast Through Time")
	}
}

// Taigam: after it attacks, an instant cast from hand gains rebound
// from the trigger and is exiled as it resolves.
func TestTaigamGivesRebound_AfterItAttacked(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	taigam := pushCatalogPermanent(g, me.ID, "Taigam, Ojutai Master", "Legendary Creature — Human Monk", taigamOracle, false)
	declareAttack(t, g, opp.ID, taigam)
	advanceTo(t, g, game.StepPostcombatMain)

	id := castPlainSpell(t, g, me, "Instant")
	if n := triggersOnStackFrom(g, taigam); n != 1 {
		t.Fatalf("Taigam triggers on the stack = %d, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Fatal("the instant Taigam gave rebound was not exiled as it resolved")
	}
	if n := reboundTriggersFor(g, id); n != 1 {
		t.Errorf("rebound triggers = %d, want 1", n)
	}
}

// The intervening "if" (CR 603.4): a Taigam that has not attacked this
// turn does not trigger, and the spell goes to the graveyard.
func TestTaigamGivesNothing_BeforeItAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	taigam := pushCatalogPermanent(g, me.ID, "Taigam, Ojutai Master", "Legendary Creature — Human Monk", taigamOracle, false)

	id := castPlainSpell(t, g, me, "Sorcery")
	if n := triggersOnStackFrom(g, taigam); n != 0 {
		t.Fatalf("Taigam triggered without attacking: %d triggers", n)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) || g.Exile.Contains(id) {
		t.Error("a sorcery cast before Taigam attacked rebounded")
	}
}

// "From your hand": a spell cast from anywhere else does not trigger
// Taigam, and creature spells never do.
func TestTaigamIgnoresCastsNotFromYourHand(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	taigam := pushCatalogPermanent(g, me.ID, "Taigam, Ojutai Master", "Legendary Creature — Human Monk", taigamOracle, false)
	declareAttack(t, g, opp.ID, taigam)
	advanceTo(t, g, game.StepPostcombatMain)

	// The instant sits where the cast would leave it; only the event's
	// origin zone differs between the two questions.
	instant := uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(game.Card{InstanceID: instant, Name: "Some Instant", TypeLine: "Instant",
			ManaCost: "{0}", Owner: me.ID, Controller: me.ID})
	})
	src := battlefieldCardPtr(t, g, taigam)
	cast := func(from game.ZoneKind) bool {
		var fired bool
		g.WithWriteLock(func() {
			fired = taigamCastFromHandAfterAttacking(game.Event{
				Kind: game.EventCast, Actor: me.ID, CardID: instant, OldZone: from, NewZone: game.ZoneStack,
			}, src, game.Characteristic{}, g)
		})
		return fired
	}
	if !cast(game.ZoneHand) {
		t.Fatal("setup: Taigam does not trigger on an instant cast from hand")
	}
	for _, from := range []game.ZoneKind{game.ZoneExile, game.ZoneGraveyard, game.ZoneCommand} {
		if cast(from) {
			t.Errorf("Taigam triggered on a cast from %s", from)
		}
	}
	g.WithWriteLock(func() { _, _ = g.Stack.Remove(instant) })
	bear := castFromHandForTest(t, g, me, "Grizzly Bears", "Creature — Bear", "{0}", "", game.CastSpellParams{})
	if n := triggersOnStackFrom(g, taigam); n != 0 {
		t.Errorf("Taigam triggered on a creature spell (%s)", bear)
	}
}

// Taigam's shield: instants, sorceries and Dragon spells you control
// can't be countered; your other spells and an opponent's instant can.
func TestTaigamShieldsInstantsSorceriesAndDragons(t *testing.T) {
	dragon := csCard("Shivan Dragon", "Creature — Dragon", "{4}{R}{R}", 5)
	for want, spells := range map[bool][]csSpell{
		true:  {{card: csBlueInstant}, {card: csSorcery}, {card: dragon}},
		false: {{card: csBear}, {card: csBlueInstant, caster: 1, controller: 1}},
	} {
		for _, s := range spells {
			g := newCatalogGame(t)
			pushCatalogPermanent(g, g.Seats[0].ID, "Taigam, Ojutai Master", "Legendary Creature — Human Monk", taigamOracle, false)
			id := csPushSpell(g, s)
			if got := counterByEffect(t, g, id); got != want {
				t.Errorf("%s (controlled by seat %d): survived the counter = %v, want %v", s.card.Name, s.controller, got, want)
			}
		}
	}
}

// battlefieldCardPtr is a pointer to the battlefield card `id`.
func battlefieldCardPtr(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	t.Fatalf("%s is not on the battlefield", id)
	return nil
}

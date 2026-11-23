package testsg

import (
	"testing"

	"github.com/google/uuid"

	cardssy "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/sy"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cast Through Time: an instant its controller casts from hand is
// exiled as it resolves and rebounds; a creature spell is not touched,
// and the opponent's instant goes to the graveyard.
func TestCastThroughTimeGivesYourInstantsAndSorceriesRebound(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Cast Through Time", "Enchantment", effectstest.CastThroughTimeOracle, false)

	instant := effectstest.CastPlainSpell(t, g, me, "Instant")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(instant) || me.Graveyard.Contains(instant) {
		t.Fatal("an instant cast from hand under Cast Through Time was not exiled by rebound")
	}
	if n := effectstest.ReboundTriggersFor(g, instant); n != 1 {
		t.Errorf("rebound triggers = %d, want 1", n)
	}

	sorcery := effectstest.CastPlainSpell(t, g, me, "Sorcery")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(sorcery) || effectstest.ReboundTriggersFor(g, sorcery) != 1 {
		t.Error("a sorcery cast from hand under Cast Through Time did not rebound")
	}

	g.WithWriteLock(func() { g.Turn.PriorityHolder = (seat + 1) % len(g.Seats) })
	theirs := effectstest.CastPlainSpell(t, g, opp, "Instant")
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's instant rebounded off your Cast Through Time")
	}
}

// Taigam: after it attacks, an instant cast from hand gains rebound
// from the trigger and is exiled as it resolves.
func TestTaigamGivesRebound_AfterItAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	taigam := effectstest.PushCatalogPermanent(g, me.ID, "Taigam, Ojutai Master", "Legendary Creature — Human Monk", effectstest.TaigamOracle, false)
	effectstest.DeclareAttack(t, g, opp.ID, taigam)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)

	id := effectstest.CastPlainSpell(t, g, me, "Instant")
	if n := effectstest.TriggersOnStackFrom(g, taigam); n != 1 {
		t.Fatalf("Taigam triggers on the stack = %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Fatal("the instant Taigam gave rebound was not exiled as it resolved")
	}
	if n := effectstest.ReboundTriggersFor(g, id); n != 1 {
		t.Errorf("rebound triggers = %d, want 1", n)
	}
}

// The intervening "if" (CR 603.4): a Taigam that has not attacked this
// turn does not trigger, and the spell goes to the graveyard.
func TestTaigamGivesNothing_BeforeItAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	taigam := effectstest.PushCatalogPermanent(g, me.ID, "Taigam, Ojutai Master", "Legendary Creature — Human Monk", effectstest.TaigamOracle, false)

	id := effectstest.CastPlainSpell(t, g, me, "Sorcery")
	if n := effectstest.TriggersOnStackFrom(g, taigam); n != 0 {
		t.Fatalf("Taigam triggered without attacking: %d triggers", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) || g.Exile.Contains(id) {
		t.Error("a sorcery cast before Taigam attacked rebounded")
	}
}

// "From your hand": a spell cast from anywhere else does not trigger
// Taigam, and creature spells never do.
func TestTaigamIgnoresCastsNotFromYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	taigam := effectstest.PushCatalogPermanent(g, me.ID, "Taigam, Ojutai Master", "Legendary Creature — Human Monk", effectstest.TaigamOracle, false)
	effectstest.DeclareAttack(t, g, opp.ID, taigam)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)

	// The instant sits where the cast would leave it; only the event's
	// origin zone differs between the two questions.
	instant := uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(game.Card{InstanceID: instant, Name: "Some Instant", TypeLine: "Instant",
			ManaCost: "{0}", Owner: me.ID, Controller: me.ID})
	})
	src := effectstest.BattlefieldCardPtr(t, g, taigam)
	cast := func(from game.ZoneKind) bool {
		var fired bool
		g.WithWriteLock(func() {
			fired = cardssy.TaigamCastFromHandAfterAttacking(game.Event{
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
	bear := effectstest.CastFromHandForTest(t, g, me, "Grizzly Bears", "Creature — Bear", "{0}", "", game.CastSpellParams{})
	if n := effectstest.TriggersOnStackFrom(g, taigam); n != 0 {
		t.Errorf("Taigam triggered on a creature spell (%s)", bear)
	}
}

// Taigam's shield: instants, sorceries and Dragon spells you control
// can't be countered; your other spells and an opponent's instant can.
func TestTaigamShieldsInstantsSorceriesAndDragons(t *testing.T) {
	dragon := effectstest.CsCard("Shivan Dragon", "Creature — Dragon", "{4}{R}{R}", 5)
	for want, spells := range map[bool][]effectstest.CsSpell{
		true:  {{Card: effectstest.CsBlueInstant}, {Card: effectstest.CsSorcery}, {Card: dragon}},
		false: {{Card: effectstest.CsBear}, {Card: effectstest.CsBlueInstant, Caster: 1, Controller: 1}},
	} {
		for _, s := range spells {
			g := effectstest.NewCatalogGame(t)
			effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Taigam, Ojutai Master", "Legendary Creature — Human Monk", effectstest.TaigamOracle, false)
			id := effectstest.CsPushSpell(g, s)
			if got := effectstest.CounterByEffect(t, g, id); got != want {
				t.Errorf("%s (controlled by seat %d): survived the counter = %v, want %v", s.Card.Name, s.Controller, got, want)
			}
		}
	}
}

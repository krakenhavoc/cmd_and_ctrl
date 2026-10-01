package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_return_together_test.go — #1867. A card that returns
// several cards from a graveyard to the battlefield puts them there in
// ONE event, so each newcomer sees the others enter (CR 603.6a: "Each
// time an event puts one or more permanents onto the battlefield, all
// permanents on the battlefield (including the newcomers) are checked
// for any enters-the-battlefield triggers that match the event").
//
// The observer is Soul Warden ("Whenever another creature enters, you
// gain 1 life"). Two Wardens returned together trigger each other: two
// triggers, 2 life. Returned one at a time, only the first sees the
// second: one trigger, 1 life.

// gtWarden puts a Soul Warden card into p's graveyard.
func gtWarden(p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Soul Warden", OracleID: b04SoulWardenOracle,
		TypeLine: "Creature — Human Cleric", Power: 1, Toughness: 1,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// gtSettle resolves the stack, answering the CR 603.3b order prompt
// two simultaneous triggers of one controller ask for.
func gtSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8; i++ {
		passPriorityAroundTable(t, g)
		if triggerOrderPrompt(g) == nil {
			return
		}
		answerTriggerOrderInOfferedOrder(t, g)
	}
	t.Fatal("the stack did not settle")
}

// Reveillark's two creature cards enter together.
func TestReveillarkReturnsItsTwoCreaturesTogether(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lark := b12Push(g, me.ID, "Reveillark", "Creature — Elemental", b34ReveillarkOracle, 4, 3)
	a, b := gtWarden(me), gtWarden(me)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(lark) })
	b04WaitForPick(t, g, me.ID)
	b17PickCards(t, g, me.ID, a, b)
	life := me.Life
	gtSettle(t, g)
	if !onBattlefield(g, a) || !onBattlefield(g, b) {
		t.Fatal("both Wardens return")
	}
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2: each Warden sees the other enter (CR 603.6a)", got)
	}
}

// Rise of the Dark Realms takes every graveyard's creatures in one
// entry, under the caster's control.
func TestRiseOfTheDarkRealmsPutsEveryGraveyardsCreaturesTogether(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a, b := gtWarden(me), gtWarden(opp)
	castCatalogSpell(t, g, "Rise of the Dark Realms", "Sorcery", b04RiseOfTheDarkRealmsOracle, nil)
	life := me.Life
	gtSettle(t, g)
	for _, id := range []uuid.UUID{a, b} {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !onBattlefield(g, id) || c.Controller != me.ID {
			t.Fatalf("%s is on the battlefield under the caster's control", id)
		}
	}
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2: each Warden sees the other enter (CR 603.6a)", got)
	}
}

// Exhume: the players choose in APNAP order, and the creatures enter
// together once the last has chosen. Each player's Warden sees the
// other player's enter, so each player gains 1.
func TestExhumePutsEveryPlayersCreatureOntoTheBattlefieldTogether(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine, theirs := gtWarden(me), gtWarden(opp)
	castCatalogSpell(t, g, "Exhume", "Sorcery", b38ExhumeOracle, nil)
	passPriorityAroundTable(t, g)
	myLife, oppLife := me.Life, opp.Life
	for _, pick := range []struct {
		seat *game.Player
		card uuid.UUID
	}{{me, mine}, {opp, theirs}} {
		ask := latestChooseCardsFor(g, pick.seat.ID)
		if ask == nil {
			t.Fatalf("%s is not asked", pick.seat.Name)
		}
		if err := g.ResolveChooseCards(ask.ID, pick.seat.ID, []uuid.UUID{pick.card}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	gtSettle(t, g)
	if !onBattlefield(g, mine) || !onBattlefield(g, theirs) {
		t.Fatal("both chosen Wardens enter")
	}
	if me.Life-myLife != 1 || opp.Life-oppLife != 1 {
		t.Errorf("life gained %d and %d, want 1 each: each Warden sees the other enter (CR 603.6a)",
			me.Life-myLife, opp.Life-oppLife)
	}
}

// Splendid Reclamation's lands enter tapped, through the tapped clause
// the batch carries on each entry event.
func TestSplendidReclamationLandsEnterTogetherAndTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := batch01GraveyardCard(me, "Forest", "Basic Land — Forest")
	b := batch01GraveyardCard(me, "Island", "Basic Land — Island")
	castCatalogSpell(t, g, "Splendid Reclamation", "Sorcery", b07SplendidReclamationOracle, nil)
	before := len(g.Events)
	passPriorityAroundTable(t, g)
	var etbs []uuid.UUID
	g.ReadSnapshot(func() {
		for _, ev := range g.Events[before:] {
			if ev.Kind == game.EventETB {
				etbs = append(etbs, ev.CardID)
			}
		}
		for _, id := range []uuid.UUID{a, b} {
			c, ok := g.LookupCardForEffect(id)
			if !ok || !onBattlefield(g, id) {
				t.Errorf("%s did not return", id)
				continue
			}
			if !c.Tapped {
				t.Errorf("%s entered untapped", c.Name)
			}
		}
	})
	if len(etbs) != 2 {
		t.Errorf("%d ETB events, want 2", len(etbs))
	}
}

// Pull's two creature cards enter together.
func TestPullPutsItsTwoCreaturesOntoTheBattlefieldTogether(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a, b := gtWarden(opp), gtWarden(opp)
	advanceToMain(t, g)
	card := mdfcCard(me.ID, oraclePushPull, game.LayoutSplit,
		game.Face{Name: "Push", TypeLine: "Sorcery", ManaCost: "{0}"},
		game.Face{Name: "Pull", TypeLine: "Sorcery", ManaCost: "{0}"},
	)
	me.Hand.PushTop(card)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: sgCards(a, b)}); err != nil {
		t.Fatalf("cast Pull: %v", err)
	}
	life := me.Life
	gtSettle(t, g)
	if !onBattlefield(g, a) || !onBattlefield(g, b) {
		t.Fatal("both Wardens enter")
	}
	// Pull's Wardens are the caster's now, so the caster gains.
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2: each Warden sees the other enter (CR 603.6a)", got)
	}
}

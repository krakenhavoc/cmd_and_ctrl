package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reveillark's two creature cards enter together.
func TestReveillarkReturnsItsTwoCreaturesTogether(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lark := effectstest.B12Push(g, me.ID, "Reveillark", "Creature — Elemental", effectstest.B34ReveillarkOracle, 4, 3)
	a, b := effectstest.GtWarden(me), effectstest.GtWarden(me)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(lark) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B17PickCards(t, g, me.ID, a, b)
	life := me.Life
	effectstest.GtSettle(t, g)
	if !effects.OnBattlefield(g, a) || !effects.OnBattlefield(g, b) {
		t.Fatal("both Wardens return")
	}
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2: each Warden sees the other enter (CR 603.6a)", got)
	}
}

// Rise of the Dark Realms takes every graveyard's creatures in one
// entry, under the caster's control.
func TestRiseOfTheDarkRealmsPutsEveryGraveyardsCreaturesTogether(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a, b := effectstest.GtWarden(me), effectstest.GtWarden(opp)
	effectstest.CastCatalogSpell(t, g, "Rise of the Dark Realms", "Sorcery", effectstest.B04RiseOfTheDarkRealmsOracle, nil)
	life := me.Life
	effectstest.GtSettle(t, g)
	for _, id := range []uuid.UUID{a, b} {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !effects.OnBattlefield(g, id) || c.Controller != me.ID {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine, theirs := effectstest.GtWarden(me), effectstest.GtWarden(opp)
	effectstest.CastCatalogSpell(t, g, "Exhume", "Sorcery", effectstest.B38ExhumeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	myLife, oppLife := me.Life, opp.Life
	for _, pick := range []struct {
		seat *game.Player
		card uuid.UUID
	}{{me, mine}, {opp, theirs}} {
		ask := effectstest.LatestChooseCardsFor(g, pick.seat.ID)
		if ask == nil {
			t.Fatalf("%s is not asked", pick.seat.Name)
		}
		if err := g.ResolveChooseCards(ask.ID, pick.seat.ID, []uuid.UUID{pick.card}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	effectstest.GtSettle(t, g)
	if !effects.OnBattlefield(g, mine) || !effects.OnBattlefield(g, theirs) {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.Batch01GraveyardCard(me, "Forest", "Basic Land — Forest")
	b := effectstest.Batch01GraveyardCard(me, "Island", "Basic Land — Island")
	effectstest.CastCatalogSpell(t, g, "Splendid Reclamation", "Sorcery", effectstest.B07SplendidReclamationOracle, nil)
	before := len(g.Events)
	effectstest.PassPriorityAroundTable(t, g)
	var etbs []uuid.UUID
	g.ReadSnapshot(func() {
		for _, ev := range g.Events[before:] {
			if ev.Kind == game.EventETB {
				etbs = append(etbs, ev.CardID)
			}
		}
		for _, id := range []uuid.UUID{a, b} {
			c, ok := g.LookupCardForEffect(id)
			if !ok || !effects.OnBattlefield(g, id) {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a, b := effectstest.GtWarden(opp), effectstest.GtWarden(opp)
	effectstest.AdvanceToMain(t, g)
	card := effectstest.MdfcCard(me.ID, effectstest.OraclePushPull, game.LayoutSplit,
		game.Face{Name: "Push", TypeLine: "Sorcery", ManaCost: "{0}"},
		game.Face{Name: "Pull", TypeLine: "Sorcery", ManaCost: "{0}"},
	)
	me.Hand.PushTop(card)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: effectstest.SgCards(a, b)}); err != nil {
		t.Fatalf("cast Pull: %v", err)
	}
	life := me.Life
	effectstest.GtSettle(t, g)
	if !effects.OnBattlefield(g, a) || !effects.OnBattlefield(g, b) {
		t.Fatal("both Wardens enter")
	}
	// Pull's Wardens are the caster's now, so the caster gains.
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2: each Warden sees the other enter (CR 603.6a)", got)
	}
}

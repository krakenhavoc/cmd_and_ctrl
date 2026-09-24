package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// departed_source_colour_test.go is the card-level proof for #1417: a
// damage source that has LEFT the battlefield is judged by the colour
// it had as it last existed there (CR 608.2h) — by protection
// (CR 702.16e) and by every "a red source you control" damage
// replacement — not by its card's colour in the graveyard.
//
// Every test plays the same real line. Murderous Redcap (black-red)
// enters and its trigger targets; in response Cerulean Wisps turns it
// blue ("becomes blue" replaces its colours, CR 105.3); then it dies,
// and its trigger deals 2 damage from the departed Redcap. That Redcap
// was BLUE. The card in the graveyard is black-red again.

// redcapTurnedBlueThenKilled casts a black-red Redcap, lets its enter
// trigger target `target`, turns it blue with Cerulean Wisps in
// response, destroys it, and resolves the trigger.
func redcapTurnedBlueThenKilled(t *testing.T, g *game.Game, target game.TargetRef) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	redcap := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: redcap, Name: "Murderous Redcap", TypeLine: "Creature — Goblin Assassin",
		OracleID: b40MurderousRedcapOracle, Power: 2, Toughness: 2, Colors: []string{"B", "R"},
		Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, redcap, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Redcap: %v", err)
	}
	passUntilOnBattlefield(t, g, redcap)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, target); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	if triggerOnStack(g, redcap) == nil {
		t.Fatal("setup: the Redcap's enter trigger should be waiting")
	}

	wisps := castCatalogSpell(t, g, "Cerulean Wisps", "Instant", b43CeruleanWispsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: redcap}})
	for i := 0; i < 8 && g.Stack.Contains(wisps); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if g.Stack.Contains(wisps) || triggerOnStack(g, redcap) == nil {
		t.Fatal("setup: the Wisps should have resolved with the Redcap trigger still waiting")
	}
	var colors []string
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, _ := g.LookupCardForEffect(redcap)
		colors = c.Effective().Colors
	})
	if len(colors) != 1 || colors[0] != "U" {
		t.Fatalf("setup: the Redcap is %v on the battlefield, want blue", colors)
	}

	removeInResponse(t, g, redcap, false)
	var inYard game.Card
	g.WithWriteLock(func() { inYard, _ = g.LookupCardForEffect(redcap) })
	if !inYard.HasColor("R") {
		t.Fatal("setup: the Redcap card in the graveyard must be red, or these tests prove nothing")
	}
	passPriorityAroundTable(t, g)
}

// Protection from blue prevents the departed BLUE Redcap's damage.
// Reading the black-red graveyard card let it through.
func TestDepartedBlueRedcapIsPreventedByProtectionFromBlue(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pro-Blue Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 5, Colors: []string{"W"}, Keywords: []string{"protection from blue"},
		Owner: opp.ID, Controller: opp.ID,
	})
	redcapTurnedBlueThenKilled(t, g, game.TargetRef{Kind: game.TargetCard, ID: victim})
	c, ok := battlefieldCard(g, victim)
	if !ok {
		t.Fatal("the pro-blue creature should still be there")
	}
	if c.DamageMarked != 0 {
		t.Errorf("pro-blue creature took %d damage from a Redcap that was blue as it "+
			"last existed (CR 608.2h / 702.16e, #1417)", c.DamageMarked)
	}
}

// Torbran adds 2 to damage from a RED source you control. The departed
// Redcap was blue, so its 2 damage is not raised.
func TestTorbranDoesNotRaiseADepartedSourceThatWasNotRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Torbran, Thane of Red Fell", "Legendary Creature — Dwarf Noble", b08TorbranOracle, false)
	before := opp.Life
	redcapTurnedBlueThenKilled(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != before-2 {
		t.Errorf("opponent lost %d, want 2: the Redcap was blue as it last existed, "+
			"so Torbran does not raise it (#1417)", before-opp.Life)
	}
}

// Ojer Axonil raises noncombat damage from a RED source you control to
// its power. The departed Redcap was blue, so 2 stays 2.
func TestOjerAxonilDoesNotRaiseADepartedSourceThatWasNotRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ojer Axonil, Deepest Might", TypeLine: "Legendary Creature — God",
		OracleID: ojerAxonilOracle, Power: 4, Toughness: 4, Colors: []string{"R"},
		Owner: me.ID, Controller: me.ID,
	})
	before := opp.Life
	redcapTurnedBlueThenKilled(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != before-2 {
		t.Errorf("opponent lost %d, want 2: the Redcap was blue as it last existed, "+
			"so Ojer does not raise it to 4 (#1417)", before-opp.Life)
	}
}

// Mechanized Warfare adds 1 to damage from a red or artifact source you
// control. The departed Redcap was a blue creature, so 2 stays 2.
func TestMechanizedWarfareDoesNotRaiseADepartedSourceThatWasNotRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Mechanized Warfare", "Enchantment", b40MechanizedWarfareOracle, false)
	before := opp.Life
	redcapTurnedBlueThenKilled(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != before-2 {
		t.Errorf("opponent lost %d, want 2: the Redcap was blue as it last existed, "+
			"so Mechanized Warfare does not raise it (#1417)", before-opp.Life)
	}
}

// The live half is unchanged: a red source still on the battlefield is
// raised by Torbran, read through the same snapshot.
func TestTorbranStillRaisesALiveRedSource(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Torbran, Thane of Red Fell", "Legendary Creature — Dwarf Noble", b08TorbranOracle, false)
	src := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Red Pinger", TypeLine: "Creature — Goblin",
		Power: 1, Toughness: 1, Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
	})
	before := opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, opp.ID, 1); err != nil {
			t.Fatalf("DealDamageToPlayerForEffect: %v", err)
		}
	})
	if opp.Life != before-3 {
		t.Errorf("opponent lost %d, want 3 (1 + Torbran's 2)", before-opp.Life)
	}
}

package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gn2_cards_test.go — the Game Night 2019 cards (#5515): Calculating Lich,
// Earthshaker Giant and Highcliff Felidar.

const (
	gn2CalculatingLichOracle  = "7264d3b3-fd46-4ea3-a85d-f0b068c331ad"
	gn2EarthshakerGiantOracle = "cd6250ae-9079-4a62-8a70-0d94fbac21bc"
	gn2HighcliffFelidarOracle = "43296f8b-58d9-446e-a538-1c4921552c41"
)

func TestGN2CardsAreRegisteredFull(t *testing.T) {
	for oracle, want := range map[string]struct {
		name    string
		keyword string
	}{
		gn2CalculatingLichOracle:  {"Calculating Lich", "menace"},
		gn2EarthshakerGiantOracle: {"Earthshaker Giant", "trample"},
		gn2HighcliffFelidarOracle: {"Highcliff Felidar", "vigilance"},
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != want.name {
			t.Errorf("%s: registered=%v name=%q", want.name, ok, spec.Name)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: want Full with no caveats, got %v / %v", want.name, spec.Completeness, spec.Caveats)
		}
		if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != want.keyword {
			t.Errorf("%s: printed keywords %v, want [%s]", want.name, spec.PrintedKeywords, want.keyword)
		}
		if len(spec.Triggered) != 1 {
			t.Errorf("%s: %d triggered abilities, want 1", want.name, len(spec.Triggered))
		}
	}
}

// --- Calculating Lich ------------------------------------------------

// gn2Lich puts Calculating Lich onto seat 0's battlefield and returns it.
func gn2Lich(g *game.Game) uuid.UUID {
	return b12Push(g, g.Seats[0].ID, "Calculating Lich", "Creature — Zombie Wizard", gn2CalculatingLichOracle, 5, 5)
}

// One opponent attacking another: each creature attacking an opponent of
// the Lich's controller costs that defender 1 life, whoever's creature it
// is.
func TestGN2CalculatingLichDrainsAnOpponentAttackedByAnother(t *testing.T) {
	g := newCatalogGame(t)
	attacker, defender := g.Seats[1], g.Seats[2]
	lich := gn2Lich(g)
	a := b12Creature(g, attacker.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, attacker.ID, "Bear B", "Creature — Bear", 2, 2)
	advanceToUpkeepOf(t, g, 1)
	life, attackerLife := defender.Life, attacker.Life
	declareAttack(t, g, defender.ID, a, b)
	if n := triggersOnStackFrom(g, lich); n != 2 {
		t.Fatalf("two creatures attacking an opponent: %d Lich triggers, want 2", n)
	}
	passPriorityAroundTable(t, g)
	if defender.Life != life-2 {
		t.Errorf("defender at %d, want %d", defender.Life, life-2)
	}
	if attacker.Life != attackerLife {
		t.Errorf("the attacking player lost life: %d → %d", attackerLife, attacker.Life)
	}
}

// Your own attackers drain the opponent they attack.
func TestGN2CalculatingLichDrainsTheOpponentYouAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lich := gn2Lich(g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMainOf(t, g, 0)
	life := opp.Life
	declareAttack(t, g, opp.ID, lich, bear)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent at %d, want %d (one per attacking creature, the Lich included)", opp.Life, life-2)
	}
}

// An attack on the Lich's controller, or on an opponent's planeswalker, is
// not an attack on one of your opponents.
func TestGN2CalculatingLichIgnoresAttacksOnYouAndOnPlaneswalkers(t *testing.T) {
	g := newCatalogGame(t)
	me, attacker, other := g.Seats[0], g.Seats[1], g.Seats[2]
	lich := gn2Lich(g)
	walker := p1g2Planeswalker(g, other.ID, 5)
	a := b12Creature(g, attacker.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, attacker.ID, "Bear B", "Creature — Bear", 2, 2)
	advanceToMainOf(t, g, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, at := range []struct{ id, def uuid.UUID }{{a, me.ID}, {b, walker}} {
		if err := g.DeclareAttacker(at.id, at.def); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	if n := triggersOnStackFrom(g, lich); n != 0 {
		t.Fatalf("attacks on you and on a planeswalker: %d Lich triggers, want 0", n)
	}
}

// --- Earthshaker Giant -----------------------------------------------

// The enters trigger pumps your other creatures and gives them trample;
// the Giant and opponents' creatures are untouched.
func TestGN2EarthshakerGiantPumpsYourOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wolf := b12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	advanceToMain(t, g)
	giant := castCatalogSpell(t, g, "Earthshaker Giant", "Creature — Giant Druid", gn2EarthshakerGiantOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(giant) {
		t.Fatal("Earthshaker Giant should be on the battlefield")
	}
	if got := effectivePower(t, g, bear); got != 5 {
		t.Errorf("your other creature's power %d, want 5", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("your other creature should have trample")
	}
	if got := effectivePower(t, g, wolf); got != 2 {
		t.Errorf("an opponent's creature's power %d, want 2", got)
	}
	if hasEffectiveKeyword(t, g, wolf, "trample") {
		t.Error("an opponent's creature gained trample")
	}
	if got := effectivePower(t, g, giant); got != 0 {
		t.Errorf("the Giant's power %d, want its test card's own 0 (it does not pump itself)", got)
	}
}

// --- Highcliff Felidar -----------------------------------------------

func gn2Creature(g *game.Game, owner uuid.UUID, name string, power int, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", Keywords: keywords,
		Power: power, Toughness: power + 1, Owner: owner, Controller: owner,
	})
}

// For each opponent the controller chooses among that player's
// greatest-power creatures: a lone greatest is taken unasked (hexproof is
// no defence, nothing targets), a tie is the controller's to break, an
// opponent with no creature is skipped, and the chosen creatures are
// destroyed together. Your own creatures are untouched.
func TestGN2HighcliffFelidarDestroysEachOpponentsGreatestPowerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	n := len(g.Seats)
	// The third opponent controls no creature and must not be asked.
	a, b := g.Seats[(g.Turn.ActiveSeat+1)%n], g.Seats[(g.Turn.ActiveSeat+2)%n]
	mine := gn2Creature(g, me.ID, "My Titan", 9)
	aBig := gn2Creature(g, a.ID, "A Giant", 5, "hexproof")
	aSmall := gn2Creature(g, a.ID, "A Bear", 2)
	bOne := gn2Creature(g, b.ID, "B Ogre One", 3)
	bTwo := gn2Creature(g, b.ID, "B Ogre Two", 3)
	bSmall := gn2Creature(g, b.ID, "B Squire", 1)
	advanceToMain(t, g)

	felidar := castAndResolveCreature(t, g, "Highcliff Felidar", "Creature — Cat Beast", gn2HighcliffFelidarOracle)
	passPriorityAroundTable(t, g)

	var legs []*game.PendingChoice
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoiceTheirPermanents && pc.Chooser == me.ID {
			legs = append(legs, pc)
		}
	}
	if len(legs) != 1 || legs[0].FromPlayer != b.ID {
		t.Fatalf("want one choice, over B's tie, answered by the Felidar's controller: %+v", legs)
	}
	leg := legs[0]
	if leg.ChooseMin != 1 || leg.ChooseMax != 1 || len(leg.ChooseCards) != 2 {
		t.Fatalf("B's leg offers %v (%d..%d), want exactly one of the two 3-power Ogres", leg.ChooseCards, leg.ChooseMin, leg.ChooseMax)
	}
	if !g.Battlefield.Contains(aBig) {
		t.Fatal("A's Giant was destroyed before the choice for B was made")
	}
	if err := g.ResolveTheirPermanents(leg.ID, me.ID, []uuid.UUID{bSmall}); err == nil {
		t.Error("a creature without the greatest power was accepted")
	}
	if err := g.ResolveTheirPermanents(leg.ID, me.ID, []uuid.UUID{bTwo}); err != nil {
		t.Fatalf("ResolveTheirPermanents: %v", err)
	}
	passPriorityAroundTable(t, g)

	for id, want := range map[uuid.UUID]bool{
		felidar: true, mine: true,
		aBig: false, aSmall: true,
		bOne: true, bTwo: false, bSmall: true,
	} {
		if got := g.Battlefield.Contains(id); got != want {
			t.Errorf("%s on the battlefield = %v, want %v", cardNameOf(g, id), got, want)
		}
	}
}

// With no tie anywhere nothing is asked, and the destruction is one event:
// an indestructible greatest creature survives while the others still go.
func TestGN2HighcliffFelidarWithoutTiesAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	n := len(g.Seats)
	a, b := g.Seats[(g.Turn.ActiveSeat+1)%n], g.Seats[(g.Turn.ActiveSeat+2)%n]
	aBig := gn2Creature(g, a.ID, "A Giant", 4, "indestructible")
	aSmall := gn2Creature(g, a.ID, "A Bear", 2)
	bOnly := gn2Creature(g, b.ID, "B Rat", 0)
	advanceToMain(t, g)

	castAndResolveCreature(t, g, "Highcliff Felidar", "Creature — Cat Beast", gn2HighcliffFelidarOracle)
	passPriorityAroundTable(t, g)

	if pc := latestChoiceOfKindFor(g, game.PendingChoiceTheirPermanents, me.ID); pc != nil {
		t.Fatalf("no tie to break, yet a choice is open: %+v", pc)
	}
	if !g.Battlefield.Contains(aBig) {
		t.Error("the indestructible Giant was destroyed")
	}
	if !g.Battlefield.Contains(aSmall) {
		t.Error("A's Bear does not have A's greatest power and was destroyed")
	}
	if g.Battlefield.Contains(bOnly) {
		t.Error("B's only creature, 0 power, is still the greatest and should be destroyed")
	}
}

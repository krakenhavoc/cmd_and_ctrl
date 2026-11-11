package effects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s36_cards_test.go — batch 36, morph payoffs. Each test plays the
// real catalog entry.

const (
	s36AinokOracle        = "7a5dae5f-29ab-4996-b236-c754de8769bc"
	s36DragonslayerOracle = "862af5c4-cb34-4bfa-a7ff-209e103defd6"
	s36DenProtectorOracle = "59324efc-65b7-4b31-89c5-17aed5c7eeb1"
	s36AmbushersOracle    = "2373e035-25f7-473a-9386-2093f77d3f94"
)

// s36CastFaceDown casts a card from the active seat's hand face down by
// its alternative cost key ("morph" or "megamorph") and resolves it.
func s36CastFaceDown(t *testing.T, g *game.Game, name, typeLine, oracle, key string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{AlternativeCost: key}); err != nil {
		t.Fatalf("cast %s face down: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	if c := findCardAnywhere(t, g, id); !c.FaceDownIsPermanent() {
		t.Fatalf("%s did not enter face down", name)
	}
	return id
}

func s36TurnUp(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, mana ...string) {
	t.Helper()
	mkmMPTurnFaceUp(t, g, p, id, mana...)
}

func TestCmdr2S36MorphCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s36AinokOracle: "Ainok Survivalist", s36DragonslayerOracle: "Hidden Dragonslayer",
		s36DenProtectorOracle: "Den Protector", s36AmbushersOracle: "Salt Road Ambushers",
	} {
		def, ok := Lookup(oracle)
		if !ok || def.Completeness != CompletenessFull {
			t.Errorf("%s: registered=%v completeness=%v", name, ok, def.Completeness)
		}
		alt := game.FaceDownCastFor(oracle)
		if alt == nil || alt.FaceDown == nil || !alt.FaceDown.FaceUpCounter {
			t.Errorf("%s has no megamorph declaration", name)
		}
	}
}

func TestCmdr2S36AinokSurvivalistDestroysAnOpponentsArtifactWhenTurnedUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	mine := b12Permanent(g, me.ID, "My Rock", "Artifact")
	id := s36CastFaceDown(t, g, "Ainok Survivalist", "Creature — Dog Shaman", s36AinokOracle, "megamorph", 2, 1)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("entering face down triggered the Survivalist")
	}
	s36TurnUp(t, g, me, id, "G", "C")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the opponent's artifact survived")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own artifact was destroyed")
	}
	if n := counterCount(g, id, game.CounterPlusOne); n != 1 {
		t.Errorf("megamorph counters = %d, want 1", n)
	}
}

func TestCmdr2S36HiddenDragonslayerDestroysAPowerFourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	small := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	id := s36CastFaceDown(t, g, "Hidden Dragonslayer", "Creature — Human Warrior", s36DragonslayerOracle, "megamorph", 2, 1)
	s36TurnUp(t, g, me, id, "W", "C", "C")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the power-4 creature survived")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the power-3 creature was destroyed")
	}
}

func TestCmdr2S36HiddenDragonslayerCannotTargetAPowerThreeCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	id := s36CastFaceDown(t, g, "Hidden Dragonslayer", "Creature — Human Warrior", s36DragonslayerOracle, "megamorph", 2, 1)
	s36TurnUp(t, g, me, id, "W", "C", "C")
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("a power-3 creature was destroyed")
	}
}

func TestCmdr2S36DenProtectorReturnsACardFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	card := b17GraveyardCard(me, "Old Spell", "Sorcery", "{1}")
	id := s36CastFaceDown(t, g, "Den Protector", "Creature — Human Warrior", s36DenProtectorOracle, "megamorph", 2, 1)
	s36TurnUp(t, g, me, id, "G", "C")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, card)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) {
		t.Error("the card did not return to hand")
	}
}

func TestCmdr2S36DenProtectorSmallerCreaturesCannotBlockIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	protector := b12Push(g, me.ID, "Den Protector", "Creature — Human Warrior", s36DenProtectorOracle, 3, 1)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	small := b12Creature(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", 2, 2)
	equal := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	brAttack(t, g, protector, bear)
	offered := brOffered(t, g, opp.ID)
	if brOffers(offered, small, protector) {
		t.Error("a power-2 creature may block the power-3 Den Protector")
	}
	if !brOffers(offered, small, bear) {
		t.Error("Den Protector's rule also stops blocks of another attacker")
	}
	if !brOffers(offered, equal, protector) {
		t.Error("a creature with equal power may not block (less than is strict)")
	}
}

func TestCmdr2S36SaltRoadAmbushersGivesTurnedUpCreaturesTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Salt Road Ambushers", "Creature — Dog Warrior", s36AmbushersOracle, 3, 3)
	id := s36CastFaceDown(t, g, "Ainok Survivalist", "Creature — Dog Shaman", s36AinokOracle, "megamorph", 2, 1)
	if n := counterCount(g, id, game.CounterPlusOne); n != 0 {
		t.Fatalf("entering face down got %d counters", n)
	}
	s36TurnUp(t, g, me, id, "G", "C")
	passPriorityAroundTable(t, g)
	// one megamorph counter plus the Ambushers' two
	if n := counterCount(g, id, game.CounterPlusOne); n != 3 {
		t.Errorf("counters = %d, want 3", n)
	}
}

func TestCmdr2S36SaltRoadAmbushersIgnoresAnOpponentsPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, opp.ID, "Salt Road Ambushers", "Creature — Dog Warrior", s36AmbushersOracle, 3, 3)
	id := s36CastFaceDown(t, g, "Ainok Survivalist", "Creature — Dog Shaman", s36AinokOracle, "megamorph", 2, 1)
	s36TurnUp(t, g, me, id, "G", "C")
	passPriorityAroundTable(t, g)
	if n := counterCount(g, id, game.CounterPlusOne); n != 1 {
		t.Errorf("counters = %d, want only the megamorph one", n)
	}
}

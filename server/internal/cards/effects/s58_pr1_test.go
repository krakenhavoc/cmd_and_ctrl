package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S58 PR 1: the review stamps (Reanimate, Rakdos Charm, Toxic Deluge,
// Vandalblast, Sakura-Tribe Elder, Imoti) and four MDFC front faces. The
// earlier tests for the six stamped cards live beside their families;
// these pin the clauses those left unproved.

const (
	s58FellTheProfaneOracle    = "053a69d8-2b5e-4f14-8b02-ca405891dc4a"
	s58BridgeworksBattleOracle = "9d581188-ce80-494e-bd38-f411e1f4efb5"
	s58StumpStompOracle        = "eb7b1284-0b2c-4b6a-a389-b2b932838083"
	s58SunderingEruptionOracle = "c95309e9-5c2f-4518-b2fd-825d3d0a4ae0"
)

func s58Refs(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

// --- Reanimate -----------------------------------------------------

// CR 202.3e: {X} in a graveyard is 0, so an {X}{R}{R} creature costs 2
// life, read off the card as it sat in the graveyard.
func TestReanimateChargesXAsZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dead := seedGraveyardCreature(g.Seats[1], "Hydra", "{X}{R}{R}")
	before := me.Life

	castCatalogSpell(t, g, "Reanimate", "Sorcery", reanimateOracle, s58Refs(dead))
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature did not return")
	}
	if got := before - me.Life; got != 2 {
		t.Errorf("life paid = %d, want 2 ({X} counts as 0 in a graveyard)", got)
	}
}

// "Target creature card": a noncreature card in a graveyard is refused
// at announce, and one that leaves before resolution fizzles the spell
// with no life paid (CR 608.2b).
func TestReanimateTargetsACreatureCardAndFizzlesWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	rock := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: rock, Name: "Sol Ring", TypeLine: "Artifact",
		ManaCost: "{1}", Owner: opp.ID, Controller: opp.ID})
	if err := castCatalogSpellErr(t, g, "Reanimate", "Sorcery", reanimateOracle, s58Refs(rock)); err == nil {
		t.Fatal("Reanimate accepted an artifact card as its target")
	}

	dead := seedGraveyardCreature(opp, "Bulky Zombie", "{4}{B}")
	before := me.Life
	castCatalogSpell(t, g, "Reanimate", "Sorcery", reanimateOracle, s58Refs(dead))
	if _, err := opp.Graveyard.Remove(dead); err != nil {
		t.Fatalf("removing the target in response: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(dead) {
		t.Error("a creature that left its graveyard still came back")
	}
	if me.Life != before {
		t.Errorf("life %d -> %d, want no payment when the spell fizzles", before, me.Life)
	}
}

// --- Toxic Deluge --------------------------------------------------

// CR 611.2c: the -X/-X is locked to the creatures there as it resolves.
// A creature that arrives later is not shrunk.
func TestToxicDelugeDoesNotShrinkALaterCreature(t *testing.T) {
	g := newCatalogGame(t)
	castWipe(t, g, "Toxic Deluge", "Sorcery", toxicDelugeOracle, game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)

	late := pushWipeCreature(g, g.Seats[1].ID, "Late Bear", "Creature — Bear", 2, 2)
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == late && c.CurrentToughness() != 2 {
			t.Errorf("a creature that arrived after the Deluge has toughness %d, want 2", c.CurrentToughness())
		}
	}
}

// --- Vandalblast ---------------------------------------------------

// "Target artifact you don't control": your own artifact is refused when
// hard cast.
func TestVandalblastRefusesYourOwnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := seedPermanentFor(g, me.ID, "My Signet", "Artifact")
	err := castCatalogSpellErr(t, g, "Vandalblast", "Sorcery", vandalblastOracle, s58Refs(mine))
	if err != game.ErrIllegalTarget {
		t.Fatalf("own artifact as the target: err = %v, want ErrIllegalTarget", err)
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the refused cast still destroyed the artifact")
	}
}

// --- Sakura-Tribe Elder --------------------------------------------

// The cost is a sacrifice and nothing else, so a summoning-sick Elder
// pays it at once; the basic arrives tapped, a nonbasic land is not
// offered, and the Elder is gone.
func TestSakuraTribeElderFetchesATappedBasicTheTurnItArrives(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elder := pushCatalogPermanent(g, me.ID, "Sakura-Tribe Elder", "Creature — Snake Shaman", sakuraTribeElderOracle, true)
	forest := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	nonbasic := stapleLibraryCard(me, "Stomping Ground", "Land — Mountain Forest")

	if err := g.ActivateCatalogAbility(me.ID, elder, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("a summoning-sick Elder could not sacrifice itself: %v", err)
	}
	if g.Battlefield.Contains(elder) {
		t.Error("the Elder is sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		if searchOptionNamed(g, c, "Stomping Ground") != uuid.Nil {
			t.Error("a nonbasic land was offered to a basic land search")
		}
		answerSearchByID(t, g, me.ID, forest)
	}
	got, ok := battlefieldCard(g, forest)
	if !ok {
		t.Fatal("the basic land is not on the battlefield")
	}
	if !got.Tapped {
		t.Error("the fetched basic must enter tapped")
	}
	if g.Battlefield.Contains(nonbasic) {
		t.Error("the nonbasic land was fetched")
	}
}

// --- Imoti ---------------------------------------------------------

// Imoti HAS cascade: casting it (mana value 5) is one cascade offer, and
// it is its own, since Imoti is on the stack and not the battlefield.
func TestImotiHasCascadeItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 6)

	castWithCost(t, g, "Imoti, Celebrant of Bounty", "Legendary Creature — Snake Druid", "{3}{G}{U}", imotiOracle)
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 1 {
		t.Errorf("casting Imoti got %d cascade offers, want 1", got)
	}
	answerAllMayCast(t, g, me.ID, false)
}

// "Spells YOU cast": an opponent's six-drop gets no cascade from your
// Imoti.
func TestImotiDoesNotGrantCascadeToAnOpponentsSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Imoti, Celebrant of Bounty", "Legendary Creature — Snake Druid", imotiOracle, false)
	advanceToMainOf(t, g, 1)
	opp := g.Seats[1]
	seedCheapLibrary(g, opp, 6)

	castWithCost(t, g, "Six Drop", "Sorcery", "{5}{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 0 {
		t.Errorf("an opponent's six-drop got %d cascade offers from your Imoti, want 0", got)
	}
}

// --- Fell the Profane ----------------------------------------------

func TestFellTheProfaneDestroysACreatureOrPlaneswalkerOnly(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	walker := pushWalkerForTest(g, opp.ID, "Their Walker", "", 4)
	land := seedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")

	if err := castCatalogSpellErr(t, g, "Fell the Profane", "Instant", s58FellTheProfaneOracle, s58Refs(land)); err == nil {
		t.Fatal("Fell the Profane accepted a land")
	}
	castCatalogSpell(t, g, "Fell the Profane", "Instant", s58FellTheProfaneOracle, s58Refs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived")
	}
	castCatalogSpell(t, g, "Fell the Profane", "Instant", s58FellTheProfaneOracle, s58Refs(walker))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the planeswalker survived")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land was destroyed")
	}
}

// --- Stump Stomp ---------------------------------------------------

func TestStumpStompIsOneSided(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Beast", "Creature — Beast", 4, 4)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	walker := pushWalkerForTest(g, opp.ID, "Their Walker", "", 5)

	castCatalogSpell(t, g, "Stump Stomp", "Sorcery", s58StumpStompOracle, s58Refs(mine, theirs))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("4 damage should kill the Bear")
	}
	if damageMarkedOn(g, mine) != 0 {
		t.Error("the Beast deals the damage and takes none")
	}
	castCatalogSpell(t, g, "Stump Stomp", "Sorcery", s58StumpStompOracle, s58Refs(mine, walker))
	passPriorityAroundTable(t, g)
	if got := loyaltyOf(g, walker); got != 1 {
		t.Errorf("walker loyalty = %d, want 1", got)
	}
	if err := castCatalogSpellErr(t, g, "Stump Stomp", "Sorcery", s58StumpStompOracle, s58Refs(mine, mine)); err != game.ErrIllegalTarget {
		t.Errorf("your own creature as the victim: err = %v, want ErrIllegalTarget", err)
	}
}

// --- Bridgeworks Battle --------------------------------------------

// +2/+2 first, then the fight, with both powers read after the pump: a
// 2/2 becomes 4/4, kills the 3/3 and survives its 3 damage.
func TestBridgeworksBattlePumpsThenFights(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)

	castCatalogSpell(t, g, "Bridgeworks Battle", "Sorcery", s58BridgeworksBattleOracle, s58Refs(mine, theirs))
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(theirs) {
		t.Error("the pumped 4/4 should kill the 3/3")
	}
	c, ok := battlefieldCard(g, mine)
	if !ok {
		t.Fatal("the 4/4 died to 3 damage")
	}
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("pumped creature is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if damageMarkedOn(g, mine) != 3 {
		t.Errorf("damage on the fighter = %d, want 3 (a fight is two-sided)", damageMarkedOn(g, mine))
	}
}

// "Up to one": with no second target the pump still happens and nothing
// fights.
func TestBridgeworksBattleWithNoVictimStillPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	bystander := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Bridgeworks Battle", "Sorcery", s58BridgeworksBattleOracle, s58Refs(mine))
	passPriorityAroundTable(t, g)

	c, ok := battlefieldCard(g, mine)
	if !ok || c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("creature = %+v, want a 4/4 on the battlefield", c)
	}
	if damageMarkedOn(g, bystander) != 0 || damageMarkedOn(g, mine) != 0 {
		t.Error("nothing should have fought")
	}
}

// The second clause is "you don't control": your own second creature is
// refused.
func TestBridgeworksBattleRefusesYourOwnVictim(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Bridgeworks Battle", "Sorcery", s58BridgeworksBattleOracle, s58Refs(a, b)); err != game.ErrIllegalTarget {
		t.Errorf("own creature as the fight target: err = %v, want ErrIllegalTarget", err)
	}
}

// --- Sundering Eruption --------------------------------------------

// Destroy, the victim's optional tapped basic, then Falter's rule. A
// flyer may still block; the land's controller is the one searched for.
func TestSunderingEruptionDestroysFetchesAndStopsGroundBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	land := seedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	forest := stapleLibraryCard(opp, "Forest", "Basic Land — Forest")
	grounded := pushSizedCreature(g, opp.ID, "Grounded Bear", 2, 2)
	bird := pushSizedCreature(g, opp.ID, "Bird", 1, 1, "flying")

	castCatalogSpell(t, g, "Sundering Eruption", "Sorcery", s58SunderingEruptionOracle, s58Refs(land))
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("the caster was offered the search; it belongs to the land's controller")
	}
	answerSearchByID(t, g, opp.ID, forest)
	got, ok := battlefieldCard(g, forest)
	if !ok || !got.Tapped {
		t.Errorf("the fetched basic must be on the battlefield tapped, got %+v (found %v)", got, ok)
	}
	assertRestrictions(t, g, grounded, game.CantBlock)
	assertRestrictions(t, g, bird, 0)
}

// The search is a "may": declining it leaves the library alone and the
// can't-block rider still applies.
func TestSunderingEruptionDeclinedSearchStillStopsBlockers(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land := seedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	forest := stapleLibraryCard(opp, "Forest", "Basic Land — Forest")
	grounded := pushSizedCreature(g, opp.ID, "Grounded Bear", 2, 2)

	castCatalogSpell(t, g, "Sundering Eruption", "Sorcery", s58SunderingEruptionOracle, s58Refs(land))
	passPriorityAroundTable(t, g)
	answerSearchFailToFind(t, g, opp.ID)

	if g.Battlefield.Contains(forest) {
		t.Error("the basic was fetched after declining")
	}
	assertRestrictions(t, g, grounded, game.CantBlock)
}

// --- the stamps ----------------------------------------------------

func TestS58PR1CardsAreFull(t *testing.T) {
	for _, oracle := range []string{
		reanimateOracle, rakdosCharmOracle, toxicDelugeOracle, vandalblastOracle,
		sakuraTribeElderOracle, imotiOracle, s58FellTheProfaneOracle,
		s58BridgeworksBattleOracle, s58StumpStompOracle, s58SunderingEruptionOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness = %s, want full", spec.Name, spec.Completeness)
		}
	}
}

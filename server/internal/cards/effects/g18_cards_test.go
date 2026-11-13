package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// g18_cards_test.go — the four M19 Gift Pack cards (#5524): Angelic
// Guardian, Angler Turtle, Immortal Phoenix and Vengeant Vampire, each
// at the four-seat catalog table.

const (
	g18AngelicGuardian = "2f959c67-2b0b-4ba8-a8aa-90c25a3806ec"
	g18AnglerTurtle    = "a09da13e-f8a3-410e-941a-9912d2cfe7de"
	g18ImmortalPhoenix = "1d8bd391-b133-4c8b-95ad-7f7ab99137c7"
	g18VengeantVampire = "e9f4d669-d915-4984-bd3c-12202c366e9b"
)

func TestG18CardsAreRegisteredFull(t *testing.T) {
	for oracle, want := range map[string]string{
		g18AngelicGuardian: "Angelic Guardian",
		g18AnglerTurtle:    "Angler Turtle",
		g18ImmortalPhoenix: "Immortal Phoenix",
		g18VengeantVampire: "Vengeant Vampire",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", want)
			continue
		}
		if spec.Name != want || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", want, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// g18Destroy destroys a permanent the way a "destroy" effect does.
func g18Destroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
}

// --- Angelic Guardian -------------------------------------------------

// One trigger for the declaration, with the Guardian itself at home;
// the creatures that attacked gain indestructible until end of turn,
// and the Guardian, a creature that stayed home and an opponent's
// creature do not.
func TestAngelicGuardianMakesTheAttackersIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	angel := b12Push(g, me.ID, "Angelic Guardian", "Creature — Angel", g18AngelicGuardian, 5, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wolf := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	home := b12Creature(g, me.ID, "Homebody", "Creature — Elf", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	if !hasKeywordOnBattlefield(t, g, angel, "flying") {
		t.Error("Angelic Guardian has no flying")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(wolf, third.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	if n := triggersOnStackFrom(g, angel); n != 1 {
		t.Fatalf("%d triggers for one declaration of two attackers, want 1", n)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, wolf} {
		if !hasKeywordOnBattlefield(t, g, id, "indestructible") {
			t.Errorf("an attacker has no indestructible")
		}
	}
	for _, id := range []uuid.UUID{angel, home, theirs} {
		if hasKeywordOnBattlefield(t, g, id, "indestructible") {
			t.Errorf("a creature that did not attack gained indestructible")
		}
	}
	g18Destroy(t, g, bear)
	if !onBattlefield(g, bear) {
		t.Fatal("an indestructible attacker was destroyed")
	}
	endTurn(t, g)
	if hasKeywordOnBattlefield(t, g, bear, "indestructible") {
		t.Error("indestructible should end with the turn")
	}
}

// "Creatures you control": an opponent's attack sets nothing off.
func TestAngelicGuardianIgnoresAnOpponentsAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := b12Push(g, me.ID, "Angelic Guardian", "Creature — Angel", g18AngelicGuardian, 5, 5)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(theirs, me.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	if n := triggersOnStackFrom(g, angel); n != 0 {
		t.Errorf("%d triggers for an opponent's attack, want 0", n)
	}
	passPriorityAroundTable(t, g)
	if hasKeywordOnBattlefield(t, g, theirs, "indestructible") {
		t.Error("an opponent's attacker gained indestructible")
	}
}

// CR 400.7: an attacker that leaves the battlefield and comes back
// before the trigger resolves is a new object and gains nothing; the
// other attacker still does.
func TestAngelicGuardianSkipsAnAttackerThatLeftAndCameBack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := b12Push(g, me.ID, "Angelic Guardian", "Creature — Angel", g18AngelicGuardian, 5, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, angel, bear)
	if n := triggersOnStackFrom(g, angel); n != 1 {
		t.Fatalf("%d triggers, want 1", n)
	}
	killCreature(t, g, me.ID, bear)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, bear); err != nil {
		t.Fatalf("MoveCardByID back to the battlefield: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, angel, "indestructible") {
		t.Error("the Guardian attacked and has no indestructible")
	}
	if hasKeywordOnBattlefield(t, g, bear, "indestructible") {
		t.Error("an attacker that left and came back gained indestructible")
	}
}

// --- Angler Turtle ----------------------------------------------------

// An opponent's creature must attack, but any player will do: at a
// four-seat table it may attack a third player rather than the
// Turtle's controller (CR 508.1d). The Turtle's controller's own
// creatures are not required to attack.
func TestAnglerTurtleMakesOpponentsCreaturesAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	turtle := b12Push(g, me.ID, "Angler Turtle", "Creature — Turtle", g18AnglerTurtle, 5, 7)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !hasKeywordOnBattlefield(t, g, turtle, "hexproof") {
		t.Error("Angler Turtle has no hexproof")
	}

	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("its controller's pass with their own creatures home: %v", err)
	}
	_ = mine

	advanceToDeclareAttackersOf(t, g, 1)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != theirs {
		t.Fatalf("opponent's pass with their Bear home = %v, want a refusal naming it", err)
	}
	if got := re.Sentence(opp.ID); got != "Their Bear must attack this combat if able (Angler Turtle)." {
		t.Errorf("sentence = %q", got)
	}
	if err := g.DeclareAttacker(theirs, third.ID); err != nil {
		t.Fatalf("DeclareAttacker at a third player: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with their Bear attacking a third player: %v", err)
	}
}

// The requirement belongs to the Turtle: with it gone, nothing has to
// attack.
func TestAnglerTurtleRequirementEndsWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	turtle := b12Push(g, me.ID, "Angler Turtle", "Creature — Turtle", g18AnglerTurtle, 5, 7)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	killCreature(t, g, me.ID, turtle)
	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Turtle gone: %v", err)
	}
}

// --- Immortal Phoenix -------------------------------------------------

func TestImmortalPhoenixReturnsToItsOwnersHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phoenix := b12Push(g, me.ID, "Immortal Phoenix", "Creature — Phoenix", g18ImmortalPhoenix, 5, 3)
	if !hasKeywordOnBattlefield(t, g, phoenix, "flying") {
		t.Error("Immortal Phoenix has no flying")
	}
	killCreature(t, g, me.ID, phoenix)
	drcSettle(t, g, me.ID)
	if !me.Hand.Contains(phoenix) || me.Graveyard.Contains(phoenix) {
		t.Fatal("the Phoenix did not return to hand")
	}

	// Stolen: it dies under an opponent's control and goes to its
	// owner's hand.
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Immortal Phoenix", TypeLine: "Creature — Phoenix",
		OracleID: g18ImmortalPhoenix, Power: 5, Toughness: 3, Owner: me.ID, Controller: opp.ID,
	})
	killCreature(t, g, me.ID, stolen)
	drcSettle(t, g, opp.ID)
	if !me.Hand.Contains(stolen) || opp.Hand.Contains(stolen) {
		t.Error("a stolen Phoenix did not go to its owner's hand")
	}
}

// Only a death: exiled, it stays in exile.
func TestImmortalPhoenixDoesNotReturnFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	phoenix := b12Push(g, me.ID, "Immortal Phoenix", "Creature — Phoenix", g18ImmortalPhoenix, 5, 3)
	exileCreature(t, g, phoenix)
	if !stackFullyEmpty(g) {
		t.Fatal("leaving for exile set the trigger off")
	}
	if me.Hand.Contains(phoenix) {
		t.Error("an exiled Phoenix returned to hand")
	}
}

// --- Vengeant Vampire -------------------------------------------------

// g18PickTarget answers the trigger's target prompt, when there is one.
func g18PickTarget(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	p := latestChoiceOfKindFor(g, game.PendingChoicePickTarget, chooser)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: target}); err != nil {
		t.Fatalf("pick target: %v", err)
	}
}

// The death destroys the chosen opponent's creature, from any opponent
// at the table, and its controller gains 4 life.
func TestVengeantVampireDestroysAnOpponentsCreatureAndGainsFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	vampire := b12Push(g, me.ID, "Vengeant Vampire", "Creature — Vampire", g18VengeantVampire, 4, 4)
	if !hasKeywordOnBattlefield(t, g, vampire, "lifelink") {
		t.Error("Vengeant Vampire has no lifelink")
	}
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	victim := b12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)
	life := me.Life

	killCreature(t, g, me.ID, vampire)
	g18PickTarget(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, victim) {
		t.Error("the targeted creature survived")
	}
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

// Its own creature is not a legal target: with no creature an opponent
// controls, the trigger is removed (CR 603.3d) and no life is gained.
func TestVengeantVampireWithNoTargetGainsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vampire := b12Push(g, me.ID, "Vengeant Vampire", "Creature — Vampire", g18VengeantVampire, 4, 4)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	life := me.Life
	killCreature(t, g, me.ID, vampire)
	drcSettle(t, g, me.ID)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("a prompt is open with no legal target: %+v", g.PendingChoices[0].Kind)
	}
	if !onBattlefield(g, mine) {
		t.Error("its controller's own creature was destroyed")
	}
	if me.Life != life {
		t.Errorf("life %d, want %d: the trigger had no target", me.Life, life)
	}
}

// Its target gone by resolution, the ability does not resolve
// (CR 608.2b): no life.
func TestVengeantVampireWithItsTargetGoneGainsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vampire := b12Push(g, me.ID, "Vengeant Vampire", "Creature — Vampire", g18VengeantVampire, 4, 4)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 2)
	life := me.Life
	killCreature(t, g, me.ID, vampire)
	g18PickTarget(t, g, me.ID, victim)
	exileCreature(t, g, victim)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d: the trigger's only target was gone", me.Life, life)
	}
}

// A legal target that is indestructible survives, and the life is
// still gained.
func TestVengeantVampireGainsLifeWhenTheTargetSurvives(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vampire := b12Push(g, me.ID, "Vengeant Vampire", "Creature — Vampire", g18VengeantVampire, 4, 4)
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Golem", TypeLine: "Creature — Golem",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"indestructible"},
	})
	b12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 2)
	life := me.Life
	killCreature(t, g, me.ID, vampire)
	g18PickTarget(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, victim) {
		t.Error("an indestructible target was destroyed")
	}
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

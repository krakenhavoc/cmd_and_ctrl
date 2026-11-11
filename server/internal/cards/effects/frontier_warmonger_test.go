package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const frontierWarmongerOracle = "9cb741ae-14d6-4a8d-bfad-703abfe3682c"

func TestFrontierWarmongerIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(frontierWarmongerOracle)
	if !ok || spec.Name != "Frontier Warmonger" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

func hasMenace(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && game.HasKeyword(&c, "menace")
}

func TestFrontierWarmongerGivesMenaceToCreaturesAttackingAnOpponentOrTheirWalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Frontier Warmonger", frontierWarmongerOracle, "Creature — Human Warrior")
	var oppWalker uuid.UUID
	g.WithWriteLock(func() {
		oppWalker = pushPermanent(g, opp.ID, game.Card{Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
			Counters: map[string]int{game.CounterLoyalty: 3}})
	})
	a := tdCreature(g, me.ID, "Attacker A", 2, 2)
	b := tdCreature(g, me.ID, "Attacker B", 2, 2)

	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(b, oppWalker); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if !hasMenace(g, a) || !hasMenace(g, b) {
		t.Errorf("menace: A=%v B=%v, want both", hasMenace(g, a), hasMenace(g, b))
	}
}

func TestFrontierWarmongerIgnoresAttacksOnItsController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Frontier Warmonger", frontierWarmongerOracle, "Creature — Human Warrior")
	raider := tdCreature(g, opp.ID, "Raider", 2, 2)
	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(raider, me.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if hasMenace(g, raider) {
		t.Error("a creature attacking the Warmonger's controller gained menace")
	}
	// An opponent's attack on a third player is still an attack on one of
	// the controller's opponents.
	g2 := newCatalogGame(t)
	me2, opp2, third2 := g2.Seats[0], g2.Seats[1], g2.Seats[2]
	pushPermanentForTest(g2, me2.ID, "Frontier Warmonger", frontierWarmongerOracle, "Creature — Human Warrior")
	r2 := tdCreature(g2, opp2.ID, "Raider", 2, 2)
	advanceToDeclareAttackersOf(t, g2, 1)
	if err := g2.DeclareAttacker(r2, third2.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g2)
	passPriorityAroundTable(t, g2)
	if !hasMenace(g2, r2) {
		t.Error("an opponent's attacker on another opponent did not gain menace")
	}
}

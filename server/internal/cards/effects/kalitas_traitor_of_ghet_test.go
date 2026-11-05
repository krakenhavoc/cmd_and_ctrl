package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kalitas_traitor_of_ghet_test.go — #4522: "If a nontoken creature an
// opponent controls would die, instead exile that card and create a 2/2
// black Zombie creature token." The token is the second instruction of
// the modified event (CR 608.2c, CR 614.6), driven through the real
// destroy, wipe, token doubler and restore paths.

const kalitasOracle = "e1cfd1cb-44a5-429f-a5c1-e6d29bad1c71"

func seedKalitas(g *game.Game, controller uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, controller, "Kalitas, Traitor of Ghet", "Legendary Creature — Vampire Warrior", kalitasOracle, false)
}

func TestKalitasIsRegisteredFullWithLifelink(t *testing.T) {
	spec, ok := Lookup(kalitasOracle)
	if !ok {
		t.Fatal("Kalitas is not registered")
	}
	if spec.Completeness != CompletenessFull || !hasKeyword(spec.PrintedKeywords, "lifelink") {
		t.Fatalf("completeness %v, keywords %v", spec.Completeness, spec.PrintedKeywords)
	}
}

// One opponent creature destroyed: it is exiled, and the Zombie is made
// at once under Kalitas' controller, with no trigger to respond to.
func TestKalitasExilesAnOpponentsCreatureAndMakesAZombie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedKalitas(g, me.ID)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if !g.Exile.Contains(bear) || opp.Graveyard.Contains(bear) {
		t.Fatalf("the bear should be exiled, not in the graveyard (exile %v)", g.Exile.Contains(bear))
	}
	if got := tokensNamedControlledBy(g, "Zombie", me.ID); got != 1 {
		t.Fatalf("%d Zombies for Kalitas' controller, want 1", got)
	}
	if len(g.PendingTriggers) != 0 || len(g.StackMeta) != 0 {
		t.Fatalf("%d pending triggers, %d stack items: the Zombie is not a trigger", len(g.PendingTriggers), len(g.StackMeta))
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie" && (c.Power != 2 || c.Toughness != 2 || !c.IsToken()) {
			t.Errorf("the Zombie is %d/%d token=%v, want a 2/2 token", c.Power, c.Toughness, c.IsToken())
		}
	}
}

// A creature that dies to lethal damage (a state-based action) is
// exiled and still makes its Zombie.
func TestKalitasOnAStateBasedDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedKalitas(g, me.ID)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if err := g.MarkDamage(bear, 2); err != nil {
		t.Fatalf("MarkDamage: %v", err)
	}
	if !g.Exile.Contains(bear) {
		t.Fatal("the creature that died to damage was not exiled")
	}
	if got := tokensNamedControlledBy(g, "Zombie", me.ID); got != 1 {
		t.Fatalf("%d Zombies, want 1", got)
	}
}

// A wipe of three opponent creatures: three exiles and three Zombies,
// all made before the wipe's continuation runs.
func TestKalitasOneZombiePerCreatureInAWipe(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedKalitas(g, me.ID)
	ids := []uuid.UUID{
		pushVanillaCreature(g, g.Seats[1].ID, "Bear A", 2, 2),
		pushVanillaCreature(g, g.Seats[2].ID, "Bear B", 2, 2),
		pushVanillaCreature(g, g.Seats[3].ID, "Bear C", 2, 2),
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentsKeyedForEffect(ids, game.KeyedTail{}); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	for _, id := range ids {
		if !g.Exile.Contains(id) {
			t.Errorf("a creature of the wipe is not in exile")
		}
	}
	if got := tokensNamedControlledBy(g, "Zombie", me.ID); got != 3 {
		t.Fatalf("%d Zombies, want 3", got)
	}
}

// Its controller's own creature, and an opponent's token, die normally
// and make no Zombie.
func TestKalitasIgnoresYourOwnCreaturesAndTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedKalitas(g, me.ID)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	if !me.Graveyard.Contains(mine) {
		t.Fatal("your own creature should go to your graveyard")
	}
	makeTokens(t, g, opp.ID, TokenCard("1/1 white Soldier"), 1)
	var soldier uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" && c.Controller == opp.ID {
			soldier = c.InstanceID
		}
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(soldier) })
	if g.Exile.Contains(soldier) {
		t.Fatal("an opponent's token was exiled; Kalitas names nontoken creatures")
	}
	if got := tokensNamedControlledBy(g, "Zombie", me.ID); got != 0 {
		t.Fatalf("%d Zombies, want none", got)
	}
}

// A token doubler under Kalitas' controller doubles the Zombie: the
// creation is an ordinary CR 701.7b event.
func TestKalitasZombieIsDoubledByParallelLives(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedKalitas(g, me.ID)
	pushTokenReplacementCard(g, parallelLivesOracle, "Parallel Lives", "Enchantment", me.ID)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if got := tokensNamedControlledBy(g, "Zombie", me.ID); got != 2 {
		t.Fatalf("%d Zombies with Parallel Lives, want 2", got)
	}
}

// Sacrificing the Zombie it made puts two +1/+1 counters on Kalitas.
func TestKalitasSacrificesAZombieForTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kalitas := seedKalitas(g, me.ID)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	var zombie uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie" && c.Controller == me.ID {
			zombie = c.InstanceID
		}
	}
	if zombie == uuid.Nil {
		t.Fatal("no Zombie to sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, kalitas, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{kalitas},
	}); err == nil {
		t.Fatal("Kalitas paid its own sacrifice; the cost names another Vampire or Zombie")
	}
	fillPoolColored(me, "B", 3)
	if err := g.ActivateCatalogAbility(me.ID, kalitas, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{zombie},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(kalitas)
	if !ok || c.Counters[game.CounterPlusOne] != 2 {
		t.Fatalf("Kalitas has %d +1/+1 counters, want 2", c.Counters[game.CounterPlusOne])
	}
}

// A restore with nothing paused: the replacement is the catalog's, so
// the restored table still exiles and makes the Zombie.
func TestKalitasSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedKalitas(g, me.ID)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	restored := restoreThroughJSON(t, g)
	rme := restored.Seats[0]
	restored.WithWriteLock(func() { _ = restored.DestroyPermanentForEffect(bear) })
	if !restored.Exile.Contains(bear) {
		t.Fatal("the restored table did not exile the creature")
	}
	if got := tokensNamedControlledBy(restored, "Zombie", rme.ID); got != 1 {
		t.Fatalf("%d Zombies on the restored table, want 1", got)
	}
}

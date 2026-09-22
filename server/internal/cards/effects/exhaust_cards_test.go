package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exhaust_cards_test.go — #1181, the catalog half of the exhaust
// keyword. The engine half (the record, its two scopes, the CR 400.7
// key, the clone and the snapshot) is in game/exhaust_test.go; the
// enumerator's and the wire's halves are in internal/legal and
// internal/protocol. Everything here is a real printed card.
//
// Four proofs, chosen for what each one adds:
//
//	Prowcatcher Specialist  the keyword and nothing else
//	Greenbelt Guardian      an exhaust ability BESIDE a repeatable one
//	Bitter Work             exhaust AND a printed activation condition,
//	                        over earthbend (#1180)
//	Ba Sing Se              an earthbend activation that is NOT exhaust

const (
	prowcatcherSpecialistOracle = "594132ac-32a7-41d5-b7f0-3692bbed7f6f"
	greenbeltGuardianOracle     = "2d8aa053-289d-40d9-baa7-9bd1c5b8e957"
	bitterWorkOracle            = "e9a24ed8-e844-4c50-b386-bc28da9989b1"
	baSingSeOracle              = "de1ae205-ca5b-4d26-8194-ca85f1406e53"
)

// fillPoolColored drops n mana of one colour into a seat's pool.
// fillPool's colourless mana cannot pay a {R} or a {G}.
func fillPoolColored(p *game.Player, color string, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(game.ManaToken{Color: color})
	}
}

// exhaustDeclared reads the Exhaust bit off a catalog card's ability
// list, which is what the keyword is.
func exhaustDeclared(t *testing.T, oracle string, index int) bool {
	t.Helper()
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: oracle})
	if index >= len(abilities) {
		t.Fatalf("oracle %s has %d activated abilities, wanted index %d", oracle, len(abilities), index)
	}
	return abilities[index].Exhaust
}

// --- Prowcatcher Specialist ------------------------------------------

// TestProwcatcherSpecialistExhaustsAfterOneActivation is the smallest
// printed exhaust card: one ability, one use, for the whole game.
func TestProwcatcherSpecialistExhaustsAfterOneActivation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	goblin := pushCatalogPermanent(g, me.ID, "Prowcatcher Specialist",
		"Creature — Goblin Warrior", prowcatcherSpecialistOracle, false)
	if !exhaustDeclared(t, prowcatcherSpecialistOracle, 0) {
		t.Fatal("the card's only ability is not declared as an exhaust ability")
	}

	fillPool(me, 3)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, goblin, game.CounterPlusOne); got != 2 {
		t.Fatalf("+1/+1 counters = %d, want 2", got)
	}

	fillPool(me, 3)
	fillPoolColored(me, "R", 1)
	err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrAbilityExhausted) {
		t.Fatalf("second activation: %v, want ErrAbilityExhausted", err)
	}
	if len(me.ManaPool) != 4 {
		t.Errorf("pool holds %d mana, want 4 — a refused activation pays nothing", len(me.ManaPool))
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, goblin, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d after the refusal, want 2", got)
	}
}

// --- Greenbelt Guardian ----------------------------------------------

// TestGreenbeltGuardianKeepsItsRepeatableAbilityAfterExhausting is the
// per-ABILITY proof on a printed card: the Elf's "{G}: target creature
// gains trample" is unaffected by spending its exhaust ability, in
// both directions.
func TestGreenbeltGuardianKeepsItsRepeatableAbilityAfterExhausting(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := pushCatalogPermanent(g, me.ID, "Greenbelt Guardian",
		"Creature — Elf Ranger", greenbeltGuardianOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if exhaustDeclared(t, greenbeltGuardianOracle, 0) {
		t.Error("the repeatable {G} pump must not be declared exhaust")
	}
	if !exhaustDeclared(t, greenbeltGuardianOracle, 1) {
		t.Fatal("the {3}{G} ability is not declared as an exhaust ability")
	}

	// The pump, twice, before the exhaust ability is touched.
	for i := 0; i < 2; i++ {
		fillPoolColored(me, "G", 1)
		if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		}); err != nil {
			t.Fatalf("pump %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
	}

	fillPool(me, 3)
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, elf, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the exhaust ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, elf, game.CounterPlusOne); got != 3 {
		t.Fatalf("+1/+1 counters = %d, want 3", got)
	}

	// The pump still works, and the exhaust ability still does not.
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Errorf("the repeatable ability after the exhaust was spent: %v", err)
	}
	passPriorityAroundTable(t, g)
	fillPool(me, 3)
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, elf, 1, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrAbilityExhausted) {
		t.Errorf("the exhaust ability again: %v, want ErrAbilityExhausted", err)
	}
}

// --- Bitter Work -----------------------------------------------------

// TestBitterWorkEarthbendsOnceAndOnlyOnYourTurn is the card #1178
// stopped at, with both of its activation gates on one board. They are
// different rules and the engine keeps them apart: the condition is
// false on someone else's turn and true again on yours, and the
// exhaust is never true again.
func TestBitterWorkEarthbendsOnceAndOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	work := pushCatalogPermanent(g, me.ID, "Bitter Work", "Enchantment", bitterWorkOracle, false)
	land := pushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	other := pushEarthbendLand(g, me.ID, "Mountain", "Basic Land — Mountain")
	if !exhaustDeclared(t, bitterWorkOracle, 0) {
		t.Fatal("Bitter Work's activated ability is not declared as an exhaust ability")
	}

	fillPool(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, work, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err != nil {
		t.Fatalf("earthbend on your own turn: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, land, game.CounterPlusOne); got != 4 {
		t.Fatalf("+1/+1 counters on the earthbent land = %d, want 4", got)
	}

	fillPool(me, 4)
	err := g.ActivateCatalogAbility(me.ID, work, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}},
	})
	if !errors.Is(err, game.ErrAbilityExhausted) {
		t.Fatalf("second activation on the same turn: %v, want ErrAbilityExhausted", err)
	}
	if got := countersOn(g, other, game.CounterPlusOne); got != 0 {
		t.Errorf("the refused activation earthbent the second land anyway (%d counters)", got)
	}
}

// TestBitterWorksConditionIsCheckedBesideItsExhaust — a FRESH Bitter
// Work on somebody else's turn is refused for the printed reason
// ("Activate only during your turn"), not as exhausted. The two gates
// are separate errors because they recover differently.
func TestBitterWorksConditionIsCheckedBesideItsExhaust(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	work := pushCatalogPermanent(g, opp.ID, "Bitter Work", "Enchantment", bitterWorkOracle, false)
	pushEarthbendLand(g, opp.ID, "Forest", "Basic Land — Forest")
	land := pushEarthbendLand(g, opp.ID, "Swamp", "Basic Land — Swamp")

	fillPool(opp, 4)
	err := g.ActivateCatalogAbility(opp.ID, work, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	})
	if !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("on someone else's turn: %v, want ErrConditionNotMet", err)
	}
	if errors.Is(err, game.ErrAbilityExhausted) {
		t.Error("a never-activated exhaust ability must not report as exhausted")
	}
}

// TestBitterWorkDrawsWhenYouAttackWithABigCreature is the card's other
// sentence: one card per player attacked with one or more creatures of
// power 4 or greater, however many of them there are (CR 603.2c).
func TestBitterWorkDrawsWhenYouAttackWithABigCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Bitter Work", "Enchantment", bitterWorkOracle, false)
	small := pushVanillaCreature(g, me.ID, "Runt", 2, 2)
	bigA := pushVanillaCreature(g, me.ID, "Brute A", 4, 4)
	bigB := pushVanillaCreature(g, me.ID, "Brute B", 5, 5)
	before := me.Hand.Size()

	// attackWith declares and then walks to the combat damage step,
	// which is where the declaration locks in and EventAttack fires.
	attackWith(t, g, opp.ID, small, bigA, bigB)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d cards, want 1 — two big attackers at one player is one occurrence", got)
	}
}

// --- Ba Sing Se ------------------------------------------------------

// TestBaSingSeEarthbendsEveryTurnBecauseItIsNotExhaust is the line
// between the two keywords: this land prints "Activate only as a
// sorcery" and nothing else, so it may earthbend again once it untaps.
func TestBaSingSeEarthbendsEveryTurnBecauseItIsNotExhaust(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	city := pushCatalogPermanent(g, me.ID, "Ba Sing Se", "Land", baSingSeOracle, false)
	first := pushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	second := pushEarthbendLand(g, me.ID, "Plains", "Basic Land — Plains")
	if exhaustDeclared(t, baSingSeOracle, 0) {
		t.Fatal("Ba Sing Se's earthbend is declared exhaust — the printed card says only " +
			"\"Activate only as a sorcery\"")
	}

	activate := func(target uuid.UUID) error {
		fillPool(me, 2)
		fillPoolColored(me, "G", 1)
		return g.ActivateCatalogAbility(me.ID, city, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		})
	}

	if err := activate(first); err != nil {
		t.Fatalf("first earthbend: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, first, game.CounterPlusOne); got != 2 {
		t.Fatalf("+1/+1 counters = %d, want 2", got)
	}

	// Untap it, which is all that stands between the land and a second
	// activation — no exhaust record does.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == city {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := activate(second); err != nil {
		t.Fatalf("second earthbend: %v, want it to be legal — this is not an exhaust ability", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, second, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters on the second land = %d, want 2", got)
	}
}

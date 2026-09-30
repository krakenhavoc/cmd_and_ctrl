package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auras_294a_layers_test.go — the slice 294-a Auras that rewrite what
// the host is (Imprisoned in the Moon, Amphibian Downpour) or grant it
// something (Shiny Impetus, Sticky Fingers).

// --- Shiny Impetus ------------------------------------------------

func TestShinyImpetusPumpsAndGoadsTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := auraBear(g, opp.ID)
	auraCast(t, g, "Shiny Impetus", shinyImpetusOracle, theirs)

	if p, tt := effectivePower(t, g, theirs), effectiveToughness(t, g, theirs); p != 4 || tt != 4 {
		t.Errorf("enchanted bear is %d/%d, want 4/4", p, tt)
	}

	// On its controller's turn the goaded creature must attack, and
	// must attack somebody other than the Aura's controller.
	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("passing with the goaded creature home = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(theirs, me.ID); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("declaring an attack on the goader while another player is open = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(theirs, third.ID); err != nil {
		t.Fatalf("re-point the attack at another player: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("attacking a player other than the goader was refused: %v", err)
	}
}

func TestShinyImpetusEndsWhenTheAuraLeaves(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := auraBear(g, opp.ID)
	aura := auraCast(t, g, "Shiny Impetus", shinyImpetusOracle, theirs)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })

	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.PassPriority(); err != nil {
		t.Errorf("the creature is still forced to attack after the Aura left: %v", err)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("power %d after the Aura left, want 2", p)
	}
}

func TestShinyImpetusMakesATreasureForItsControllerWhenTheCreatureAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := auraBear(g, me.ID)
	aura := auraCast(t, g, "Shiny Impetus", shinyImpetusOracle, bear)

	declareAttack(t, g, opp.ID, bear)
	if triggerOnStack(g, aura) == nil {
		t.Fatal("the attack put no Shiny Impetus trigger on the stack")
	}
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Treasure", me.ID); n != 1 {
		t.Errorf("%d Treasure tokens, want 1", n)
	}
}

// --- Sticky Fingers -----------------------------------------------

func TestStickyFingersGivesMenaceAndATreasureOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := auraBear(g, me.ID)
	auraCast(t, g, "Sticky Fingers", stickyFingersOracle, bear)

	if !containsString(effectiveAbilities(t, g, bear), "menace") {
		t.Errorf("abilities %v lack menace", effectiveAbilities(t, g, bear))
	}
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Treasure", me.ID); n != 1 {
		t.Errorf("%d Treasure tokens after combat damage, want 1", n)
	}
}

// The Treasure trigger is the CREATURE's, so a stolen creature pays its
// current controller; the draw is the AURA's, so its controller draws.
func TestStickyFingersSplitsTheTreasureAndTheDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	auraSeedLibrary(me, 3)
	auraSeedLibrary(opp, 3)
	theirs := auraBear(g, opp.ID)
	auraCast(t, g, "Sticky Fingers", stickyFingersOracle, theirs)

	dealCombatDamageToPlayer(g, theirs, me.ID, 2)
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Treasure", opp.ID); n != 1 {
		t.Errorf("the creature's controller has %d Treasures, want 1", n)
	}
	if n := countOnBattlefield(g, "Treasure", me.ID); n != 0 {
		t.Errorf("the Aura's controller got %d Treasures, want 0", n)
	}

	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - myHand; got != 1 {
		t.Errorf("the Aura's controller drew %d, want 1", got)
	}
	if got := opp.Hand.Size() - theirHand; got != 0 {
		t.Errorf("the creature's controller drew %d, want 0", got)
	}
}

// --- Imprisoned in the Moon ---------------------------------------

func TestImprisonedInTheMoonTurnsACreatureIntoAColorlessLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flying Bear", TypeLine: "Creature — Bear",
		Colors: []string{"G"}, Keywords: []string{"flying"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	aura := auraCast(t, g, "Imprisoned in the Moon", imprisonedInTheMoonOracle, bear)

	types := effectiveTypes(t, g, bear)
	if len(types) != 1 || types[0] != "Land" {
		t.Errorf("types %v, want only Land", types)
	}
	if subs := effectiveSubtypes(t, g, bear); len(subs) != 0 {
		t.Errorf("subtypes %v, want none", subs)
	}
	if containsString(effectiveAbilities(t, g, bear), "flying") {
		t.Errorf("the land kept flying: %v", effectiveAbilities(t, g, bear))
	}
	var colors []string
	g.ReadSnapshot(func() { c, _ := g.LookupCardForEffect(bear); colors = c.Effective().Colors })
	if len(colors) != 0 {
		t.Errorf("colors %v, want colorless", colors)
	}
	if n := grantedManaRowsOf(g, bear); n != 1 {
		t.Fatalf("%d granted mana abilities, want exactly one", n)
	}
	if err := tapGrantedMana(t, g, me.ID, bear, ""); err != nil {
		t.Fatalf("tap the land: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("mana pool %v, want one {C}", got)
	}
	if host := attachmentHostOf(t, g, aura); host.ID != bear {
		t.Errorf("the Aura fell off a permanent that is still a legal host: %+v", host)
	}
}

func TestImprisonedInTheMoonStripsALandOfItsTypeAndItsOwnMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: opp.ID, Controller: opp.ID,
	})
	auraCast(t, g, "Imprisoned in the Moon", imprisonedInTheMoonOracle, forest)

	if containsString(effectiveSubtypes(t, g, forest), "Forest") {
		t.Errorf("the enchanted land is still a Forest: %v", effectiveSubtypes(t, g, forest))
	}
	if err := tapGrantedMana(t, g, opp.ID, forest, ""); err != nil {
		t.Fatalf("tap the land: %v", err)
	}
	if got := batch01PoolColors(opp); len(got) != 1 || got[0] != "C" {
		t.Errorf("mana pool %v, want {C} and not {G}", got)
	}
	_ = me
}

func TestImprisonedInTheMoonEndsWhenTheAuraLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	aura := auraCast(t, g, "Imprisoned in the Moon", imprisonedInTheMoonOracle, bear)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })

	if !containsString(effectiveTypes(t, g, bear), "Creature") {
		t.Errorf("types %v, want the Creature back", effectiveTypes(t, g, bear))
	}
	if n := grantedManaRowsOf(g, bear); n != 0 {
		t.Errorf("%d granted mana abilities remain", n)
	}
}

// --- Amphibian Downpour -------------------------------------------

func TestAmphibianDownpourMakesABlueFrog(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	golem := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem",
		Keywords: []string{"flying"}, Colors: []string{"R"},
		Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
	})
	auraCast(t, g, "Amphibian Downpour", amphibianDownpourOracle, golem)

	if p, tt := effectivePower(t, g, golem), effectiveToughness(t, g, golem); p != 1 || tt != 1 {
		t.Errorf("enchanted creature is %d/%d, want 1/1", p, tt)
	}
	types := effectiveTypes(t, g, golem)
	if len(types) != 1 || types[0] != "Creature" {
		t.Errorf("types %v, want only Creature (it loses Artifact)", types)
	}
	if subs := effectiveSubtypes(t, g, golem); len(subs) != 1 || subs[0] != "Frog" {
		t.Errorf("subtypes %v, want only Frog", subs)
	}
	if containsString(effectiveAbilities(t, g, golem), "flying") {
		t.Errorf("it kept flying: %v", effectiveAbilities(t, g, golem))
	}
	var colors []string
	g.ReadSnapshot(func() { c, _ := g.LookupCardForEffect(golem); colors = c.Effective().Colors })
	if len(colors) != 1 || colors[0] != "U" {
		t.Errorf("colors %v, want blue", colors)
	}
}

func TestAmphibianDownpourCanBeCastOutsideAMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := auraBear(g, opp.ID)
	advanceTo(t, g, game.StepDeclareAttackers)
	castFromSeat(t, g, me, "Amphibian Downpour", auraTypeLine, amphibianDownpourOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, theirs); p != 1 {
		t.Errorf("power %d, want 1: flash lets it resolve in combat", p)
	}
}

// Storm copies the Aura once per spell cast before it. Each copy is
// its own spell with its own target and, as it resolves, becomes a
// token Aura attached to that target.
func TestAmphibianDownpourStormCopiesEnchantAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := auraBear(g, opp.ID)
	second := auraBear(g, opp.ID)
	stormFiller(t, g)

	castCatalogSpell(t, g, "Amphibian Downpour", auraTypeLine, amphibianDownpourOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: first}})
	for i := 0; i < 64 && !(stackFullyEmpty(g) && latestPickTarget(g, me.ID) == nil); i++ {
		if p := latestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: second}); err != nil {
				t.Fatalf("re-target the copy: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}

	for name, id := range map[string]uuid.UUID{"original target": first, "copy's new target": second} {
		if p, tt := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 1 || tt != 1 {
			t.Errorf("%s is %d/%d, want 1/1", name, p, tt)
		}
		if !containsString(effectiveSubtypes(t, g, id), "Frog") {
			t.Errorf("%s is not a Frog: %v", name, effectiveSubtypes(t, g, id))
		}
	}
}

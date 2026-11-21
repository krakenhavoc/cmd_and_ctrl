package testsa

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auras_294a_layers_test.go — the slice 294-a Auras that rewrite what
// the host is (Imprisoned in the Moon, Amphibian Downpour) or grant it
// something (Shiny Impetus, Sticky Fingers).

// --- Shiny Impetus ------------------------------------------------

func TestShinyImpetusPumpsAndGoadsTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := effectstest.AuraBear(g, opp.ID)
	effectstest.AuraCast(t, g, "Shiny Impetus", shinyImpetusOracle, theirs)

	if p, tt := effectstest.EffectivePower(t, g, theirs), effectstest.EffectiveToughness(t, g, theirs); p != 4 || tt != 4 {
		t.Errorf("enchanted bear is %d/%d, want 4/4", p, tt)
	}

	// On its controller's turn the goaded creature must attack, and
	// must attack somebody other than the Aura's controller.
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.AuraBear(g, opp.ID)
	aura := effectstest.AuraCast(t, g, "Shiny Impetus", shinyImpetusOracle, theirs)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })

	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	if err := g.PassPriority(); err != nil {
		t.Errorf("the creature is still forced to attack after the Aura left: %v", err)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("power %d after the Aura left, want 2", p)
	}
}

func TestShinyImpetusMakesATreasureForItsControllerWhenTheCreatureAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Shiny Impetus", shinyImpetusOracle, bear)

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if effectstest.TriggerOnStack(g, aura) == nil {
		t.Fatal("the attack put no Shiny Impetus trigger on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Treasure", me.ID); n != 1 {
		t.Errorf("%d Treasure tokens, want 1", n)
	}
}

// --- Sticky Fingers -----------------------------------------------

func TestStickyFingersGivesMenaceAndATreasureOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Sticky Fingers", stickyFingersOracle, bear)

	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "menace") {
		t.Errorf("abilities %v lack menace", effectstest.EffectiveAbilities(t, g, bear))
	}
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Treasure", me.ID); n != 1 {
		t.Errorf("%d Treasure tokens after combat damage, want 1", n)
	}
}

// The Treasure trigger is the CREATURE's, so a stolen creature pays its
// current controller; the draw is the AURA's, so its controller draws.
func TestStickyFingersSplitsTheTreasureAndTheDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AuraSeedLibrary(me, 3)
	effectstest.AuraSeedLibrary(opp, 3)
	theirs := effectstest.AuraBear(g, opp.ID)
	effectstest.AuraCast(t, g, "Sticky Fingers", stickyFingersOracle, theirs)

	effectstest.DealCombatDamageToPlayer(g, theirs, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Treasure", opp.ID); n != 1 {
		t.Errorf("the creature's controller has %d Treasures, want 1", n)
	}
	if n := effectstest.CountOnBattlefield(g, "Treasure", me.ID); n != 0 {
		t.Errorf("the Aura's controller got %d Treasures, want 0", n)
	}

	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - myHand; got != 1 {
		t.Errorf("the Aura's controller drew %d, want 1", got)
	}
	if got := opp.Hand.Size() - theirHand; got != 0 {
		t.Errorf("the creature's controller drew %d, want 0", got)
	}
}

// --- Imprisoned in the Moon ---------------------------------------

func TestImprisonedInTheMoonTurnsACreatureIntoAColorlessLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flying Bear", TypeLine: "Creature — Bear",
		Colors: []string{"G"}, Keywords: []string{"flying"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	aura := effectstest.AuraCast(t, g, "Imprisoned in the Moon", imprisonedInTheMoonOracle, bear)

	types := effectstest.EffectiveTypes(t, g, bear)
	if len(types) != 1 || types[0] != "Land" {
		t.Errorf("types %v, want only Land", types)
	}
	if subs := effectstest.EffectiveSubtypes(t, g, bear); len(subs) != 0 {
		t.Errorf("subtypes %v, want none", subs)
	}
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "flying") {
		t.Errorf("the land kept flying: %v", effectstest.EffectiveAbilities(t, g, bear))
	}
	var colors []string
	g.ReadSnapshot(func() { c, _ := g.LookupCardForEffect(bear); colors = c.Effective().Colors })
	if len(colors) != 0 {
		t.Errorf("colors %v, want colorless", colors)
	}
	if n := effectstest.GrantedManaRowsOf(g, bear); n != 1 {
		t.Fatalf("%d granted mana abilities, want exactly one", n)
	}
	if err := effectstest.TapGrantedMana(t, g, me.ID, bear, ""); err != nil {
		t.Fatalf("tap the land: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("mana pool %v, want one {C}", got)
	}
	if host := effectstest.AttachmentHostOf(t, g, aura); host.ID != bear {
		t.Errorf("the Aura fell off a permanent that is still a legal host: %+v", host)
	}
}

func TestImprisonedInTheMoonStripsALandOfItsTypeAndItsOwnMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.AuraCast(t, g, "Imprisoned in the Moon", imprisonedInTheMoonOracle, forest)

	if effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, forest), "Forest") {
		t.Errorf("the enchanted land is still a Forest: %v", effectstest.EffectiveSubtypes(t, g, forest))
	}
	if err := effectstest.TapGrantedMana(t, g, opp.ID, forest, ""); err != nil {
		t.Fatalf("tap the land: %v", err)
	}
	if got := effectstest.Batch01PoolColors(opp); len(got) != 1 || got[0] != "C" {
		t.Errorf("mana pool %v, want {C} and not {G}", got)
	}
	_ = me
}

func TestImprisonedInTheMoonEndsWhenTheAuraLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Imprisoned in the Moon", imprisonedInTheMoonOracle, bear)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })

	if !effectstest.ContainsString(effectstest.EffectiveTypes(t, g, bear), "Creature") {
		t.Errorf("types %v, want the Creature back", effectstest.EffectiveTypes(t, g, bear))
	}
	if n := effectstest.GrantedManaRowsOf(g, bear); n != 0 {
		t.Errorf("%d granted mana abilities remain", n)
	}
}

// --- Amphibian Downpour -------------------------------------------

func TestAmphibianDownpourMakesABlueFrog(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	golem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem",
		Keywords: []string{"flying"}, Colors: []string{"R"},
		Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.AuraCast(t, g, "Amphibian Downpour", amphibianDownpourOracle, golem)

	if p, tt := effectstest.EffectivePower(t, g, golem), effectstest.EffectiveToughness(t, g, golem); p != 1 || tt != 1 {
		t.Errorf("enchanted creature is %d/%d, want 1/1", p, tt)
	}
	types := effectstest.EffectiveTypes(t, g, golem)
	if len(types) != 1 || types[0] != "Creature" {
		t.Errorf("types %v, want only Creature (it loses Artifact)", types)
	}
	if subs := effectstest.EffectiveSubtypes(t, g, golem); len(subs) != 1 || subs[0] != "Frog" {
		t.Errorf("subtypes %v, want only Frog", subs)
	}
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, golem), "flying") {
		t.Errorf("it kept flying: %v", effectstest.EffectiveAbilities(t, g, golem))
	}
	var colors []string
	g.ReadSnapshot(func() { c, _ := g.LookupCardForEffect(golem); colors = c.Effective().Colors })
	if len(colors) != 1 || colors[0] != "U" {
		t.Errorf("colors %v, want blue", colors)
	}
}

func TestAmphibianDownpourCanBeCastOutsideAMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.AuraBear(g, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.CastFromSeat(t, g, me, "Amphibian Downpour", effectstest.AuraTypeLine, amphibianDownpourOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, theirs); p != 1 {
		t.Errorf("power %d, want 1: flash lets it resolve in combat", p)
	}
}

// Storm copies the Aura once per spell cast before it. Each copy is
// its own spell with its own target and, as it resolves, becomes a
// token Aura attached to that target.
func TestAmphibianDownpourStormCopiesEnchantAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := effectstest.AuraBear(g, opp.ID)
	second := effectstest.AuraBear(g, opp.ID)
	effectstest.StormFiller(t, g)

	effectstest.CastCatalogSpell(t, g, "Amphibian Downpour", effectstest.AuraTypeLine, amphibianDownpourOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: first}})
	for i := 0; i < 64 && !(effectstest.StackFullyEmpty(g) && effectstest.LatestPickTarget(g, me.ID) == nil); i++ {
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
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
		if p, tt := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 1 || tt != 1 {
			t.Errorf("%s is %d/%d, want 1/1", name, p, tt)
		}
		if !effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, id), "Frog") {
			t.Errorf("%s is not a Frog: %v", name, effectstest.EffectiveSubtypes(t, g, id))
		}
	}
}

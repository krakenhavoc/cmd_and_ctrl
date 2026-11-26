package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b404_cards_test.go — Dragon's Maze pump and combat tricks, S82 OLD
// batch 404 (#5564): seven cards, each driven through the engine.

const (
	dgmBeetleformMage = "227cbba1-700b-4255-b730-b324ab3dec34"
	dgmWeaponSurge    = "2c46bc05-1133-4eaf-8865-eb8a1d880bd9"
	dgmDrownInFilth   = "59af03ba-28d0-4f15-b877-2cf3c81d66e1"
	dgmNivixCyclops   = "68b8bfac-63f7-4dbd-bb2e-4ad394c77335"
	dgmAetherling     = "b980bf11-9e12-4ee9-83c1-45706d9ae344"
	dgmKorozdaGorgon  = "e9e18a5a-d265-456f-a049-fa4b7e223263"
	dgmMawObzedat     = "ed8ec935-3359-42a1-8ac6-cac6f0e7f905"
)

func TestS82OldB404EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dgmBeetleformMage: "Beetleform Mage", dgmWeaponSurge: "Weapon Surge", dgmDrownInFilth: "Drown in Filth",
		dgmNivixCyclops: "Nivix Cyclops", dgmAetherling: "Aetherling", dgmKorozdaGorgon: "Korozda Gorgon",
		dgmMawObzedat: "Maw of the Obzedat",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func dgm404PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c := effectstest.LayeredCard(t, g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

func TestS82OldB404BeetleformMagePumpsAndFliesOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mage := effectstest.B21Push(g, me.ID, "Beetleform Mage", "Creature — Human Insect Wizard", dgmBeetleformMage, 2, 2, "G", "U")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	effectstest.B06AddMana(me, "G", "U", "G", "U")
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := dgm404PT(t, g, mage); p != 4 || th != 4 {
		t.Errorf("%d/%d, want 4/4", p, th)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mage, "flying") {
		t.Error("no flying")
	}
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated twice in one turn")
	}
}

func TestS82OldB404WeaponSurgeTargetedAndOverloaded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Weapon Surge", "Instant", dgmWeaponSurge, effectstest.CardTarget(theirs)); err == nil {
		t.Fatal("an opponent's creature was targeted")
	}
	effectstest.CastCatalogSpell(t, g, "Weapon Surge", "Instant", dgmWeaponSurge, effectstest.CardTarget(a))
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := dgm404PT(t, g, a); p != 3 {
		t.Errorf("targeted creature power %d, want 3", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, a, "first strike") {
		t.Error("no first strike")
	}
	if p, _ := dgm404PT(t, g, b); p != 2 {
		t.Error("an untargeted creature changed")
	}
	// Overloaded: each creature you control, and only yours.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.AdvanceToMain(t, g2)
	c := effectstest.B12Creature(g2, me2.ID, "Bear C", "Creature — Bear", 2, 2)
	d := effectstest.B12Creature(g2, me2.ID, "Bear D", "Creature — Bear", 2, 2)
	theirs2 := effectstest.B12Creature(g2, opp2.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := effectstest.HandCardFull(me2, "Weapon Surge", "Instant", "{R}", dgmWeaponSurge, []string{"R"})
	if err := g2.CastSpell(me2.ID, id, game.CastSpellParams{AlternativeCost: "overload"}); err != nil {
		t.Fatalf("cast overloaded: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	for _, x := range []uuid.UUID{c, d} {
		if p, _ := dgm404PT(t, g2, x); p != 3 {
			t.Errorf("an overloaded creature has power %d, want 3", p)
		}
		if !effectstest.HasEffectiveKeyword(t, g2, x, "first strike") {
			t.Error("an overloaded creature has no first strike")
		}
	}
	if p, _ := dgm404PT(t, g2, theirs2); p != 2 {
		t.Error("an opponent's creature changed")
	}
}

func TestS82OldB404DrownInFilthCountsLandsAfterTheMill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	me.Library.Cards = nil
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	// The top two of the four are lands; one more land is already in the graveyard.
	me.Library.Cards[2].TypeLine, me.Library.Cards[3].TypeLine = "Creature — Bear", "Creature — Bear"
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Drown in Filth", "Sorcery", dgmDrownInFilth, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != 0 {
		t.Fatalf("library %d, want 0 after milling four", got)
	}
	// Three land cards in the graveyard: the old one and two milled.
	if p, th := dgm404PT(t, g, ogre); p != 2 || th != 2 {
		t.Errorf("the creature is %d/%d, want 2/2 (-3/-3)", p, th)
	}
}

func TestS82OldB404NivixCyclopsAttacksAsThoughNoDefenderAfterASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cy := effectstest.B21Push(g, me.ID, "Nivix Cyclops", "Creature — Cyclops", dgmNivixCyclops, 1, 4, "U", "R")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := dgm404PT(t, g, cy); p != 1 {
		t.Fatalf("a creature spell pumped it to %d", p)
	}
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", m404Shock, effectstest.CardTarget(effectstest.B12Creature(g, g.Seats[1].ID, "Target", "Creature — Bear", 1, 1)))
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := dgm404PT(t, g, cy); p != 4 {
		t.Errorf("power %d, want 4 after an instant", p)
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(cy, g.Seats[1].ID); err != nil {
		t.Errorf("it could not attack as though it didn't have defender: %v", err)
	}
}

func TestS82OldB404NivixCyclopsCannotAttackWithoutASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cy := effectstest.B21Push(g, me.ID, "Nivix Cyclops", "Creature — Cyclops", dgmNivixCyclops, 1, 4, "U", "R")
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(cy, g.Seats[1].ID); err == nil {
		t.Error("it attacked with defender and no spell cast")
	}
}

const m404Shock = "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d"

func TestS82OldB404KorozdaGorgonRemovesACounterToShrinkATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gorgon := effectstest.B21Push(g, me.ID, "Korozda Gorgon", "Creature — Gorgon", dgmKorozdaGorgon, 2, 5, "B", "G")
	bear := effectstest.B12Creature(g, me.ID, "Counter Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 2) })
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, gorgon, "deathtouch") {
		t.Error("no deathtouch")
	}
	effectstest.B06AddMana(me, "C", "C")
	params := game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{bear},
		Targets: effectstest.CardTarget(ogre)}
	if err := g.ActivateCatalogAbility(me.ID, gorgon, 0, params); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters left, want 1", got)
	}
	if p, th := dgm404PT(t, g, ogre); p != 2 || th != 2 {
		t.Errorf("the target is %d/%d, want 2/2", p, th)
	}
	// No counters, no activation.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	gorgon2 := effectstest.B21Push(g2, me2.ID, "Korozda Gorgon", "Creature — Gorgon", dgmKorozdaGorgon, 2, 5, "B", "G")
	bare := effectstest.B12Creature(g2, me2.ID, "Bare Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g2)
	effectstest.B06AddMana(me2, "C", "C")
	if err := g2.ActivateCatalogAbility(me2.ID, gorgon2, 0, game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{bare},
		Targets: effectstest.CardTarget(bare)}); err == nil {
		t.Error("activated with no counter to remove")
	}
}

func TestS82OldB404MawOfTheObzedatSacrificesToPumpTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	maw := effectstest.B21Push(g, me.ID, "Maw of the Obzedat", "Creature — Thrull", dgmMawObzedat, 3, 3, "W", "B")
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, maw, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without naming a creature to sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, maw, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if p, th := dgm404PT(t, g, maw); p != 4 || th != 4 {
		t.Errorf("the Maw is %d/%d, want 4/4", p, th)
	}
	if p, _ := dgm404PT(t, g, other); p != 3 {
		t.Errorf("my other creature has power %d, want 3", p)
	}
	if p, _ := dgm404PT(t, g, theirs); p != 2 {
		t.Error("an opponent's creature was pumped")
	}
}

func TestS82OldB404AetherlingFourAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ae := effectstest.B21Push(g, me.ID, "Aetherling", "Creature — Shapeshifter", dgmAetherling, 4, 5, "U")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "U", "U")
	// {1}: +1/-1, {1}: -1/+1.
	if err := g.ActivateCatalogAbility(me.ID, ae, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{1}: +1/-1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, ae, 3, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{1}: -1/+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, ae, 3, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{1}: -1/+1 again: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := dgm404PT(t, g, ae); p != 3 || th != 6 {
		t.Errorf("%d/%d, want 3/6 (+1/-1, -1/+1, -1/+1)", p, th)
	}
	// {U}: can't be blocked this turn.
	if err := g.ActivateCatalogAbility(me.ID, ae, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{U}: unblockable: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RestrictionsOf(t, g, ae)&game.CantBeBlocked == 0 {
		t.Error("it can still be blocked")
	}
	// {U}: exile, return at the next end step as a new object.
	if err := g.ActivateCatalogAbility(me.ID, ae, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{U}: flicker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if zone, _ := effectstest.CardWhere(g, me, ae); zone != "exile" {
		t.Fatalf("it is in %s, want exile", zone)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Aetherling"); n != 1 {
		t.Errorf("%d Aetherlings on the battlefield at the end step, want 1", n)
	}
}

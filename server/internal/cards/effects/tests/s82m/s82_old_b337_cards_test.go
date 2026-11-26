package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b337_cards_test.go — OLD batch 337 (Fate Reforged: manifest, #5552):
// seven cards, each played through the engine as the catalog registers it.

const (
	oldB337RageformOracle   = "2a108899-76b4-4f85-8a59-91838fc26f4a"
	oldB337QarsiOracle      = "3bcb5ad4-818e-4f10-aaff-188d713b160b"
	oldB337InvocationOracle = "4e75fc29-68e3-41a9-9f59-adb4460663b2"
	oldB337WriteOracle      = "4fc69345-4c1f-4113-8e1c-15b8e2e597fc"
	oldB337BeastOracle      = "75c81ce4-ca1d-464a-bb39-e568f8ebf71f"
	oldB337WildcallOracle   = "d96ac790-428b-4a64-8dbd-6baa73eb6210"
	oldB337NurturingOracle  = "ddbe6d9b-ea69-449c-a765-b3c16ff60c08"
)

func TestOldB337EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB337RageformOracle:   "Rageform",
		oldB337QarsiOracle:      "Qarsi High Priest",
		oldB337InvocationOracle: "Fierce Invocation",
		oldB337WriteOracle:      "Write into Being",
		oldB337BeastOracle:      "Arashin War Beast",
		oldB337WildcallOracle:   "Wildcall",
		oldB337NurturingOracle:  "Formless Nurturing",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name=%q completeness=%v caveats=%v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB337In(z *game.Zone, id uuid.UUID) bool {
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// oldB337Bear is a library card the manifest tests look for.
var oldB337Bear = game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2}

// oldB337Manifested returns the face-down permanent that is now the card `id`.
func oldB337Manifested(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	if !oldB337In(g.Battlefield, id) {
		t.Fatalf("card %v is not on the battlefield", id)
	}
	c := effectstest.GaLayered(t, g, id)
	if !c.FaceDown {
		t.Fatalf("the card is not face down")
	}
	return c
}

func TestOldB337RageformGivesDoubleStrikeToTheManifestedCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	m := effectstest.PushLibraryCard(me, oldB337Bear)
	aura := effectstest.CastCatalogSpell(t, g, "Rageform", "Enchantment", oldB337RageformOracle, nil)
	effectstest.AdResolveSpell(t, g, aura)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, aura)
	if !c.IsAura() || !c.IsAttachedTo(m) {
		t.Fatalf("Rageform aura=%v attached=%v, want an Aura on the manifested card", c.IsAura(), c.IsAttachedTo(m))
	}
	mc := oldB337Manifested(t, g, m)
	if mc.CurrentPower() != 2 || mc.CurrentToughness() != 2 || !mc.IsCreature() {
		t.Errorf("manifested card %d/%d, want a 2/2 creature", mc.CurrentPower(), mc.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g, m, "double strike") {
		t.Error("enchanted creature lacks double strike")
	}
}

func TestOldB337RageformWithAnEmptyLibraryIsPutIntoTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	aura := effectstest.CastCatalogSpell(t, g, "Rageform", "Enchantment", oldB337RageformOracle, nil)
	effectstest.AdResolveSpell(t, g, aura)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if !oldB337In(me.Graveyard, aura) {
		t.Error("an Aura attached to nothing stayed on the battlefield")
	}
}

func TestOldB337QarsiHighPriestSacrificesAnotherCreatureToManifest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	m := effectstest.PushLibraryCard(me, oldB337Bear)
	priest := effectstest.PushCatalogPermanent(g, me.ID, "Qarsi High Priest", "Creature — Human Cleric", oldB337QarsiOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
	effectstest.Acrb1AddMana(t, g, me.ID, "{B}{B}")
	effectstest.P1g3Activate(t, g, priest, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if oldB337In(g.Battlefield, fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if c := effectstest.GaLayered(t, g, priest); !c.Tapped {
		t.Error("Qarsi High Priest was not tapped for its cost")
	}
	mc := oldB337Manifested(t, g, m)
	if mc.CurrentPower() != 2 || mc.CurrentToughness() != 2 || !mc.IsCreature() {
		t.Errorf("manifested card %d/%d, want a 2/2 creature", mc.CurrentPower(), mc.CurrentToughness())
	}
}

func TestOldB337QarsiHighPriestCannotSacrificeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	priest := effectstest.PushCatalogPermanent(g, me.ID, "Qarsi High Priest", "Creature — Human Cleric", oldB337QarsiOracle, false)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Swamp", "Basic Land — Swamp", "", false)
	effectstest.Acrb1AddMana(t, g, me.ID, "{B}{B}")
	for _, victim := range []uuid.UUID{priest, land} {
		if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{victim}}); err == nil {
			t.Errorf("sacrificed %v", victim)
		}
	}
}

// oldB337CastAndResolve casts a sorcery from hand and lets it resolve.
func oldB337CastAndResolve(t *testing.T, g *game.Game, name, oracle string) {
	t.Helper()
	spell := effectstest.CastCatalogSpell(t, g, name, "Sorcery", oracle, nil)
	_ = spell
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB337ManifestThenCounters(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		want         int
	}{
		{"Fierce Invocation", oldB337InvocationOracle, 2},
		{"Formless Nurturing", oldB337NurturingOracle, 1},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		me.Library.Cards = nil
		m := effectstest.PushLibraryCard(me, oldB337Bear)
		oldB337CastAndResolve(t, g, tc.name, tc.oracle)
		mc := oldB337Manifested(t, g, m)
		if n := mc.Counters[game.CounterPlusOne]; n != tc.want {
			t.Errorf("%s: %d +1/+1 counters, want %d", tc.name, n, tc.want)
		}
		if mc.CurrentPower() != 2+tc.want || mc.CurrentToughness() != 2+tc.want {
			t.Errorf("%s: manifested card is %d/%d, want %d/%d", tc.name, mc.CurrentPower(), mc.CurrentToughness(), 2+tc.want, 2+tc.want)
		}
	}
}

func TestOldB337ManifestThenCountersWithAnEmptyLibraryDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	before := len(g.Battlefield.Cards)
	oldB337CastAndResolve(t, g, "Fierce Invocation", oldB337InvocationOracle)
	if len(g.Battlefield.Cards) != before {
		t.Errorf("the battlefield changed with an empty library: %d → %d permanents", before, len(g.Battlefield.Cards))
	}
}

func TestOldB337WildcallPutsXCounters(t *testing.T) {
	for _, x := range []int{0, 3} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		me.Library.Cards = nil
		m := effectstest.PushLibraryCard(me, oldB337Bear)
		effectstest.Pr8CastX(t, g, "Wildcall", "Sorcery", oldB337WildcallOracle, x, nil)
		effectstest.PassPriorityAroundTable(t, g)
		mc := oldB337Manifested(t, g, m)
		if n := mc.Counters[game.CounterPlusOne]; n != x {
			t.Errorf("X=%d: %d +1/+1 counters, want %d", x, n, x)
		}
		if mc.CurrentPower() != 2+x {
			t.Errorf("X=%d: power %d, want %d", x, mc.CurrentPower(), 2+x)
		}
	}
}

func TestOldB337WriteIntoBeingManifestsOneAndPlacesTheOther(t *testing.T) {
	for _, bottom := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		me.Library.Cards = nil
		deep := effectstest.PushLibraryCard(me, game.Card{Name: "Deep", TypeLine: "Land", Power: 0})
		second := effectstest.PushLibraryCard(me, game.Card{Name: "Second", TypeLine: "Instant", ManaCost: "{R}"})
		top := effectstest.PushLibraryCard(me, oldB337Bear)
		oldB337CastAndResolve(t, g, "Write into Being", oldB337WriteOracle)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
		if pick == nil {
			t.Fatal("no choose-cards prompt for the pair")
		}
		if !effectstest.HasID(pick.ChooseCards, top) || !effectstest.HasID(pick.ChooseCards, second) || effectstest.HasID(pick.ChooseCards, deep) || len(pick.ChooseCards) != 2 {
			t.Fatalf("the pair offered is %v, want the top two cards", pick.ChooseCards)
		}
		// Manifest the instant, to show a non-permanent card is legal (CR 701.40a).
		if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{second}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		mc := oldB337Manifested(t, g, second)
		if mc.CurrentPower() != 2 || mc.CurrentToughness() != 2 || !mc.IsCreature() {
			t.Errorf("manifested card %d/%d, want a 2/2 creature", mc.CurrentPower(), mc.CurrentToughness())
		}
		ask := effectstest.PutInLibraryChoiceFor(g, me.ID)
		if ask == nil || ask.LibraryPlacement != game.LibraryPlaceTopOrBottom {
			t.Fatalf("no top-or-bottom choice for the other card: %+v", ask)
		}
		var onBottom, onTop []uuid.UUID
		if bottom {
			onBottom = []uuid.UUID{top}
		} else {
			onTop = []uuid.UUID{top}
		}
		if err := g.ResolvePutInLibrary(ask.ID, me.ID, onBottom, onTop); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
		lib := me.Library.Cards
		if bottom {
			if lib[0].InstanceID != top || lib[len(lib)-1].InstanceID != deep {
				t.Errorf("the other card is not on the bottom")
			}
		} else if lib[len(lib)-1].InstanceID != top {
			t.Errorf("the other card is not on the top")
		}
		if oldB337In(g.Battlefield, top) {
			t.Error("the other card was manifested too")
		}
	}
}

func TestOldB337WriteIntoBeingWithOneCardManifestsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	m := effectstest.PushLibraryCard(me, oldB337Bear)
	oldB337CastAndResolve(t, g, "Write into Being", oldB337WriteOracle)
	oldB337Manifested(t, g, m)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Error("a pick was asked with a single card to look at")
	}
}

func TestOldB337ArashinWarBeastManifestsWhenItDamagesABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Library.Cards = nil
	m := effectstest.PushLibraryCard(me, oldB337Bear)
	beast := effectstest.PushCatalogPermanent(g, me.ID, "Arashin War Beast", "Creature — Beast", oldB337BeastOracle, false)
	blocker := effectstest.PushCatalogPermanent(g, opp.ID, "Wall", "Creature — Wall", "", false)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, beast).Power, effectstest.FindBattlefieldCardForTest(g, beast).Toughness = 6, 6
		effectstest.FindBattlefieldCardForTest(g, blocker).Power, effectstest.FindBattlefieldCardForTest(g, blocker).Toughness = 0, 4
	})
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, beast, blocker)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	oldB337Manifested(t, g, m)
}

func TestOldB337ArashinWarBeastIgnoresUnblockedDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Library.Cards = nil
	m := effectstest.PushLibraryCard(me, oldB337Bear)
	beast := effectstest.PushCatalogPermanent(g, me.ID, "Arashin War Beast", "Creature — Beast", oldB337BeastOracle, false)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, beast).Power = 6
		effectstest.FindBattlefieldCardForTest(g, beast).Toughness = 6
	})
	effectstest.DeclareAttack(t, g, opp.ID, beast)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB337In(g.Battlefield, m) {
		t.Error("the beast manifested after damaging a player, not a blocker")
	}
	if !oldB337In(me.Library, m) {
		t.Error("the top card left the library")
	}
}

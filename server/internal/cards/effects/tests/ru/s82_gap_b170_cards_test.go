package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 GAP batch 170 (Jumpstart 2022 enchantments and instants, #5948).

const (
	gapB170AuthorityOracle = "3d50266f-5a2f-4a48-ab15-87e733624fd4"
	gapB170ValorOracle     = "56a60fcb-f9b7-438f-9221-eaca8fb2649b"
	gapB170WeightOracle    = "42989c6d-75d9-4ddd-bb9f-f18c0eb005fe"
	gapB170DeathWindOracle = "a636b62a-340b-444d-9161-f6d3367b745c"
	gapB170GiftsOracle     = "162bcc9d-6edf-4b3b-83e6-2c933e2a5967"
	gapB170FungalOracle    = "04a8858e-9033-403e-9346-5a2ad1bdf680"
	gapB170FeastOracle     = "042e0533-faf4-475e-be7b-438d23c6e605"
	gapB170HungerOracle    = "296dee6b-6ec1-4ea4-8dc3-23080dd37646"
)

func TestGapB170CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB170AuthorityOracle: "Unquestioned Authority", gapB170ValorOracle: "Valor in Akros",
		gapB170WeightOracle: "Weight of Conscience", gapB170DeathWindOracle: "Death Wind",
		gapB170GiftsOracle: "Demonic Gifts", gapB170FungalOracle: "Fungal Infection",
		gapB170FeastOracle: "Ghoul's Feast", gapB170HungerOracle: "Hunger of the Howlpack",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, name %q, completeness %v, caveats %v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestGapB170UnquestionedAuthorityDrawsAndGrantsProtectionFromCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	before := effectstest.HandSize(me)
	effectstest.CastCatalogSpell(t, g, "Unquestioned Authority", "Enchantment — Aura", gapB170AuthorityOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me); got != before+1 {
		t.Errorf("hand size %d, want %d (the Aura left the hand, one card was drawn)", got, before+1)
	}
	found := false
	for _, q := range effectstest.ProtectionsOn(t, g, bear) {
		if q == "creatures" {
			found = true
		}
	}
	if !found {
		t.Errorf("protections = %v, want protection from creatures", effectstest.ProtectionsOn(t, g, bear))
	}
}

func TestGapB170ValorInAkrosPumpsEveryCreatureWhenOneEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	old := effectstest.B12Creature(g, me.ID, "Old Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Valor in Akros", "Enchantment", gapB170ValorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, old); p != 2 || tough != 2 {
		t.Fatalf("the Enchantment itself pumped: %d/%d", p, tough)
	}
	newcomer := effectstest.B12PlayFromHand(t, g, "New Bear", "Creature — Bear", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, old); p != 3 || tough != 3 {
		t.Errorf("the creature already in play is %d/%d, want 3/3", p, tough)
	}
	// The seeded newcomer has no printed P/T, so +1/+1 makes it 1/1 (and it
	// would die at 0/0 without the pump).
	if p, tough := effectstest.PtOf(t, g, newcomer); p != 1 || tough != 1 {
		t.Errorf("the creature that entered is %d/%d, want 1/1", p, tough)
	}
}

func gapB170WeightSetup(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura := effectstest.CastCatalogSpell(t, g, "Weight of Conscience", "Enchantment — Aura", gapB170WeightOracle, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, victim, aura
}

func TestGapB170WeightOfConscienceStopsAttacksAndExilesForTwoSharingCreatures(t *testing.T) {
	g, me, victim, aura := gapB170WeightSetup(t)
	effectstest.AssertRestrictions(t, g, victim, game.CantAttack)
	a := effectstest.B12Creature(g, me.ID, "Elf One", "Creature — Elf Druid", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "Elf Two", "Creature — Elf Warrior", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOf(g, victim); z != game.ZoneExile {
		t.Errorf("the enchanted creature is in %v, want exile", z)
	}
	for _, id := range []uuid.UUID{a, b} {
		if c, _ := effectstest.BattlefieldCard(g, id); !c.Tapped {
			t.Errorf("creature %v was not tapped for the cost", id)
		}
	}
}

func TestGapB170WeightOfConscienceRefusesTwoCreaturesThatShareNoType(t *testing.T) {
	g, me, victim, aura := gapB170WeightSetup(t)
	a := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b}}); err == nil {
		t.Fatal("an Elf and a Goblin paid for the ability")
	}
	if !effectstest.BattlefieldHas(g, victim) {
		t.Error("the enchanted creature left the battlefield")
	}
	for _, id := range []uuid.UUID{a, b} {
		if c, _ := effectstest.BattlefieldCard(g, id); c.Tapped {
			t.Errorf("a refused activation tapped %v", id)
		}
	}
	g.ReadSnapshot(func() {
		spec, _ := effects.Lookup(gapB170WeightOracle)
		cost := spec.Activated[0].Cost.TapOthers
		if g.TapOthersPayable(me.ID, aura, cost) {
			t.Error("the cost reads as payable with no two creatures sharing a type")
		}
	})
}

func gapB170CastX(t *testing.T, g *game.Game, oracle, name, cost string, mana string, x int, target uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushCatalogHandCard(me, name, "Instant", oracle)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == id {
				me.Hand.Cards[i].ManaCost = cost
			}
		}
	})
	effectstest.AdvanceToMainForTest(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardTarget(target), XValue: x}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapB170DeathWindShrinksByX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Big Bear", "Creature — Bear", 4, 4)
	gapB170CastX(t, g, gapB170DeathWindOracle, "Death Wind", "{X}{B}", "{B}{B}{B}", 2, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the creature is %d/%d, want 2/2", p, tough)
	}
}

func TestGapB170DeathWindKillsAtLethalX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	gapB170CastX(t, g, gapB170DeathWindOracle, "Death Wind", "{X}{B}", "{B}{B}{B}", 2, bear)
	g.RunStateChecksForTest()
	if effectstest.BattlefieldHas(g, bear) {
		t.Error("a 2/2 survived -2/-2")
	}
}

func TestGapB170DemonicGiftsPumpsAndReturnsTheCreatureUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Demonic Gifts", "Instant", gapB170GiftsOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 2 {
		t.Fatalf("the creature is %d/%d, want 4/2", p, tough)
	}
	attacker := effectstest.B12Creature(g, opp.ID, "Killer", "Creature — Ogre", 3, 3)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(attacker, bear, 2); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, bear)
	if !ok {
		t.Fatalf("the creature is in %v, want the battlefield", effectstest.ZoneOf(g, bear))
	}
	if c.Tapped {
		t.Error("the creature returned tapped")
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the returned creature is %d/%d, want a fresh 2/2", p, tough)
	}
}

func TestGapB170FungalInfectionShrinksAndMakesASaproling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Fungal Infection", "Instant", gapB170FungalOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the creature is %d/%d, want 2/2", p, tough)
	}
	if n := effectstest.B28TokensNamed(g, me.ID, "Saproling"); n != 1 {
		t.Errorf("%d Saprolings, want 1", n)
	}
}

func TestGapB170GhoulsFeastCountsCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdGraveyardCreature(me, "Dead One")
	effectstest.AdGraveyardCreature(me, "Dead Two")
	effectstest.PushGraveyardCardTyped(me, "Old Spell", "Instant")
	effectstest.CastCatalogSpell(t, g, "Ghoul's Feast", "Instant", gapB170FeastOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the creature is %d/%d, want 4/2", p, tough)
	}
}

func TestGapB170HungerOfTheHowlpackPutsOneCounterOrThreeWithMorbid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Hunger of the Howlpack", "Instant", gapB170HungerOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Fatalf("without morbid the creature is %d/%d, want 3/3", p, tough)
	}
	// A creature dies this turn: three counters instead.
	doomed := effectstest.B12Creature(g, g.Seats[1].ID, "Doomed", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(bear, doomed, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.CastCatalogSpell(t, g, "Hunger of the Howlpack", "Instant", gapB170HungerOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("with morbid the creature is %d/%d, want 6/6", p, tough)
	}
}

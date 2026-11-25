package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// enter_as_copy_static_test.go — #6262: a permanent's static "creatures
// you control enter as a copy of …" (CR 614.1c, CR 707.2). Essence of the
// Wild's "of this creature" and Infinite Reflection's "Nontoken creatures
// you control enter as a copy of enchanted creature", wired onto test
// oracles through the catalog hook the registry fills.

const (
	testEssenceOracle    = "test-6262-essence-of-the-wild"
	testReflectionOracle = "test-6262-infinite-reflection"
)

func withEnterAsCopyStatics(t *testing.T) {
	t.Helper()
	prev := game.CatalogReplacements
	t.Cleanup(func() { game.CatalogReplacements = prev })
	game.CatalogReplacements = func(key string) []game.ReplacementEffect {
		switch key {
		case testEssenceOracle:
			return []game.ReplacementEffect{effects.CreaturesYouControlEnterAsCopyOfThis("Essence: enters as a copy")}
		case testReflectionOracle:
			return []game.ReplacementEffect{effects.NontokenCreaturesYouControlEnterAsCopyOfEnchanted("Reflection: enters as a copy")}
		}
		return prev(key)
	}
}

func seedEssence(g *game.Game, controller uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Essence of the Wild", TypeLine: "Creature — Avatar",
		OracleID: testEssenceOracle, ManaCost: "{3}{G}{G}{G}", Power: 6, Toughness: 6, PrintedPTKnown: true,
		Owner: controller, Controller: controller,
	})
}

// castBear casts a vanilla 2/2 from `p`'s hand and resolves it.
func castBear(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
			PrintedPTKnown: true, Owner: p.ID, Controller: p.ID})
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Grizzly Bears: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func soldiers(t *testing.T, g *game.Game, controller uuid.UUID, n int) []uuid.UUID {
	t.Helper()
	var made []uuid.UUID
	g.WithWriteLock(func() {
		var err error
		made, err = g.CreateTokensForEffect(controller, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier",
			Power: 1, Toughness: 1, PrintedPTKnown: true}, n, game.TokenEntryOptions{})
		if err != nil {
			t.Fatalf("CreateTokensForEffect: %v", err)
		}
	})
	return made
}

// TestEssenceOfTheWildMakesGrizzlyBearsEnterAsAnEssence is the acceptance:
// with Essence of the Wild, a Grizzly Bears enters as a 6/6 Essence of the
// Wild, which has the static too; a token does as well.
func TestEssenceOfTheWildMakesGrizzlyBearsEnterAsAnEssence(t *testing.T) {
	withEnterAsCopyStatics(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seedEssence(g, me.ID)

	bear := castBear(t, g, me)
	c := effectstest.DcCard(t, g, bear)
	if c.Name != "Essence of the Wild" || c.Effective().Power != 6 || c.Effective().Toughness != 6 {
		t.Fatalf("Grizzly Bears entered as %q %d/%d, want a 6/6 Essence of the Wild", c.Name, c.Effective().Power, c.Effective().Toughness)
	}
	if c.OracleID != testEssenceOracle {
		t.Errorf("the copy's catalog key = %q, want Essence's, so it has Essence's static (CR 707.2)", c.OracleID)
	}
	for _, id := range soldiers(t, g, me.ID, 2) {
		if tc := effectstest.DcCard(t, g, id); tc.Name != "Essence of the Wild" || !tc.IsToken() {
			t.Errorf("a Soldier token entered as %q (token %v), want a token Essence of the Wild", tc.Name, tc.IsToken())
		}
	}
}

// TestEssenceOfTheWildLeavesOtherPlayersCreaturesAlone: "creatures YOU
// control" — an opponent's creature enters as itself.
func TestEssenceOfTheWildLeavesOtherPlayersCreaturesAlone(t *testing.T) {
	withEnterAsCopyStatics(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seedEssence(g, opp.ID)
	bear := castBear(t, g, me)
	if c := effectstest.DcCard(t, g, bear); c.Name != "Grizzly Bears" {
		t.Errorf("my creature entered as %q under an opponent's Essence, want itself", c.Name)
	}
}

// TestInfiniteReflectionCopiesTheEnchantedCreatureForNontokens: nontoken
// creatures enter as a copy of the enchanted creature; tokens enter as
// themselves; an unattached Aura copies nothing.
func TestInfiniteReflectionCopiesTheEnchantedCreatureForNontokens(t *testing.T) {
	withEnterAsCopyStatics(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	giant := effectstest.SeedCopyableCreature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	aura := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Infinite Reflection", TypeLine: "Enchantment — Aura",
		OracleID: testReflectionOracle, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AttachForTest(t, g, aura, giant)

	bear := castBear(t, g, me)
	if c := effectstest.DcCard(t, g, bear); c.Name != "Hill Giant" || c.Effective().Power != 3 {
		t.Errorf("a nontoken creature entered as %q, want a copy of the enchanted Hill Giant", c.Name)
	}
	for _, id := range soldiers(t, g, me.ID, 1) {
		if tc := effectstest.DcCard(t, g, id); tc.Name != "Soldier" {
			t.Errorf("a token entered as %q, want itself: the clause says nontoken", tc.Name)
		}
	}
}

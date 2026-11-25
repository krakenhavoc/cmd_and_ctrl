package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// most_common_color_test.go — #6207: "the most common color among all
// permanents or tied for most common", counted over the layered colours
// (CR 105.2).

const (
	mccDjinnOracle = "test-6207-red-djinn"
	mccAuraOracle  = "test-6207-defiance-aura"
)

func mccRegister(t *testing.T) {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: mccDjinnOracle,
		Name:     "Test Red Djinn",
		Static:   []game.StaticAbility{effects.SelfPumpWhileColorIsMostCommon("R", -2, -2)},
	})
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: mccAuraOracle,
		Name:     "Test Defiance Aura",
		Static:   []game.StaticAbility{effects.PumpEnchantedUnlessSharesAMostCommonColor(3, 3)},
	})
}

func mccPush(g *game.Game, controller uuid.UUID, name, typeLine, oracle string, colors ...string) uuid.UUID {
	return effectstest.ApaPush(g, controller, controller, game.Card{
		Name: name, TypeLine: typeLine, OracleID: oracle, Colors: colors, Power: 2, Toughness: 2,
	})
}

func mccDjinn(g *game.Game, controller uuid.UUID) uuid.UUID {
	return effectstest.ApaPush(g, controller, controller, game.Card{
		Name: "Test Red Djinn", TypeLine: "Creature — Djinn", OracleID: mccDjinnOracle,
		Colors: []string{"R"}, ManaCost: "{2}{R}{R}", Power: 6, Toughness: 5,
	})
}

func mccRecompute(g *game.Game) {
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
}

// Halam Djinn's acceptance: a 6/5 with more blue than red permanents, a
// 4/3 when red ties.
func TestSelfPumpWhileColorIsMostCommonReadsTies(t *testing.T) {
	mccRegister(t)
	g, me, opp := effectstest.P7Table(t)
	djinn := mccDjinn(g, me.ID)
	if p, tough := effectstest.EffectivePTOf(t, g, djinn); p != 4 || tough != 3 {
		t.Fatalf("alone, red is the most common: %d/%d, want 4/3", p, tough)
	}
	mccPush(g, opp.ID, "Blue One", "Creature — Test", "", "U")
	mccPush(g, opp.ID, "Blue Two", "Creature — Test", "", "U")
	mccRecompute(g)
	if p, tough := effectstest.EffectivePTOf(t, g, djinn); p != 6 || tough != 5 {
		t.Fatalf("two blue, one red: %d/%d, want 6/5", p, tough)
	}
	mccPush(g, me.ID, "Red Gold", "Artifact", "", "R", "W")
	mccRecompute(g)
	if p, tough := effectstest.EffectivePTOf(t, g, djinn); p != 4 || tough != 3 {
		t.Fatalf("red ties blue at two: %d/%d, want 4/3", p, tough)
	}
}

// Colourless permanents count for nothing, and with no coloured
// permanent no colour is common.
func TestMostCommonColorsIgnoresColorless(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	mccPush(g, me.ID, "Rock", "Artifact", "")
	var most []string
	g.ReadSnapshot(func() { most = g.MostCommonColorsForEffect(uuid.Nil, false) })
	if len(most) != 0 {
		t.Fatalf("most common with only colourless permanents = %v, want none", most)
	}
	mccPush(g, me.ID, "Green", "Creature — Test", "", "G")
	mccPush(g, me.ID, "Black", "Creature — Test", "", "B")
	g.ReadSnapshot(func() { most = g.MostCommonColorsForEffect(uuid.Nil, false) })
	if len(most) != 2 || most[0] != "B" || most[1] != "G" {
		t.Fatalf("most common = %v, want [B G]", most)
	}
}

// Call to Arms counts one player's nontoken permanents, and "isn't tied".
func TestColorIsMostCommonUntiedAmongNontoken(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	mccPush(g, opp.ID, "Their White", "Creature — Test", "", "W")
	mccPush(g, opp.ID, "Their White Two", "Creature — Test", "", "W")
	mccPush(g, me.ID, "My Red", "Creature — Test", "", "R")
	var ok bool
	g.ReadSnapshot(func() { ok = effects.ColorIsMostCommonUntiedAmongNontoken(g, me.ID, "R") })
	if !ok {
		t.Fatal("my only coloured nontoken permanent is red: want red most common untied")
	}
	mccPush(g, me.ID, "Blue Token", "Token Creature — Test", "", "U")
	g.ReadSnapshot(func() { ok = effects.ColorIsMostCommonUntiedAmongNontoken(g, me.ID, "R") })
	if !ok {
		t.Fatal("a token does not count toward the tie")
	}
	mccPush(g, me.ID, "My Blue", "Creature — Test", "", "U")
	g.ReadSnapshot(func() { ok = effects.ColorIsMostCommonUntiedAmongNontoken(g, me.ID, "R") })
	if ok {
		t.Fatal("red tied with blue: want not most common untied")
	}
}

// Barrin's Unmaking / Tsabo's Assassin: a permanent sharing a colour
// tied for most common qualifies; a colourless one never does.
func TestSharesAMostCommonColorCountsTies(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	green := mccPush(g, opp.ID, "Green", "Creature — Test", "", "G")
	mccPush(g, me.ID, "Black", "Creature — Test", "", "B")
	mccPush(g, me.ID, "Black Two", "Creature — Test", "", "B")
	rock := mccPush(g, opp.ID, "Rock", "Artifact", "")
	var g1, r1 bool
	g.ReadSnapshot(func() {
		g1 = effects.SharesAMostCommonColor(g, effectstest.FindBattlefieldCardForTest(g, green))
		r1 = effects.SharesAMostCommonColor(g, effectstest.FindBattlefieldCardForTest(g, rock))
	})
	if g1 || r1 {
		t.Fatalf("black leads: green shares=%v, colourless shares=%v, want both false", g1, r1)
	}
	mccPush(g, opp.ID, "Green Two", "Creature — Test", "", "G")
	g.ReadSnapshot(func() {
		g1 = effects.SharesAMostCommonColor(g, effectstest.FindBattlefieldCardForTest(g, green))
	})
	if !g1 {
		t.Fatal("green tied with black: want the green creature to share a most common colour")
	}
}

// Heroic Defiance: +3/+3 unless the enchanted creature shares a most
// common colour.
func TestPumpEnchantedUnlessSharesAMostCommonColor(t *testing.T) {
	mccRegister(t)
	g, me, opp := effectstest.P7Table(t)
	bear := mccPush(g, me.ID, "White Bear", "Creature — Bear", "", "W")
	aura := mccPush(g, me.ID, "Test Defiance Aura", "Enchantment — Aura", mccAuraOracle, "W")
	effectstest.AttachForTest(t, g, aura, bear)
	mccRecompute(g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Fatalf("white is the most common and the bear is white: %d/%d, want 2/2", p, tough)
	}
	for i := 0; i < 3; i++ {
		mccPush(g, opp.ID, "Blue", "Creature — Test", "", "U")
	}
	mccRecompute(g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 5 {
		t.Fatalf("blue leads: %d/%d, want 5/5", p, tough)
	}
}

package testss82v

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrifice_alt_cost_cards_test.go — #1727's proof cards, end to end
// against the real catalog: Lava Dart (flashback, one Mountain),
// Fireblast (from hand, two Mountains) and Demon of Death's Gate
// (from hand, 6 life AND three black creatures). Dread Return has its
// own file.

const (
	lavaDartOracle          = "e48891e3-30a2-4fc8-a858-cec33c6e4ab5"
	fireblastOracle         = "9dd7f27a-e862-47d7-9158-034cf4d353b8"
	demonOfDeathsGateOracle = "be465805-1a37-4212-acf1-5dfec890b338"
)

// Lava Dart: deals 1 from hand, then again from the graveyard for a
// Mountain — and the graveyard cast exiles it. An Island is not a
// Mountain.
func TestLavaDartFlashesBackForAMountain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	at := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	before := opp.Life

	dart := effectstest.CastCatalogSpell(t, g, "Lava Dart", "Instant", lavaDartOracle, at)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(dart) {
		t.Fatalf("Lava Dart cast from hand did not reach the graveyard")
	}

	island := effectstest.SeedLand(g, me.ID, "Island", "Basic Land — Island", "")
	if err := effectstest.CastForAltCost(g, dart, "graveyard", "flashback", []uuid.UUID{island}, at); err == nil {
		t.Fatalf("an Island paid \"Sacrifice a Mountain\"")
	}
	if !g.Battlefield.Contains(island) || !me.Graveyard.Contains(dart) {
		t.Fatalf("the refused flashback paid something anyway")
	}

	mountain := effectstest.SeedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	if err := effectstest.CastForAltCost(g, dart, "graveyard", "flashback", []uuid.UUID{mountain}, at); err != nil {
		t.Fatalf("flashback for a Mountain: %v", err)
	}
	if g.Battlefield.Contains(mountain) {
		t.Errorf("the Mountain survived the flashback cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("opponent lost %d, want 2 — one from hand, one flashed back", before-opp.Life)
	}
	if !g.Exile.Contains(dart) {
		t.Errorf("a flashed-back Lava Dart was not exiled")
	}
}

// Fireblast for two Mountains and no mana at all; a Mountain and an
// Island are refused with nothing paid.
func TestFireblastSacrificesTwoMountainsInsteadOfMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	at := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	before := opp.Life
	blast := effectstest.HandCardAtMain(t, g, "Fireblast", "Instant", fireblastOracle)
	m1 := effectstest.SeedLand(g, me.ID, "Mountain A", "Basic Land — Mountain", "")
	island := effectstest.SeedLand(g, me.ID, "Island", "Basic Land — Island", "")

	if err := effectstest.CastForAltCost(g, blast, "", "sacrifice", []uuid.UUID{m1, island}, at); err == nil {
		t.Fatalf("a Mountain and an Island paid \"sacrifice two Mountains\"")
	}
	if !g.Battlefield.Contains(m1) || !g.Battlefield.Contains(island) || !me.Hand.Contains(blast) {
		t.Fatalf("the refused cast paid something anyway")
	}

	m2 := effectstest.SeedLand(g, me.ID, "Mountain B", "Basic Land — Mountain", "")
	if err := effectstest.CastForAltCost(g, blast, "", "sacrifice", []uuid.UUID{m1, m2}, at); err != nil {
		t.Fatalf("Fireblast for two Mountains: %v", err)
	}
	if g.Battlefield.Contains(m1) || g.Battlefield.Contains(m2) {
		t.Errorf("a Mountain survived the cost")
	}
	if !g.Battlefield.Contains(island) {
		t.Errorf("the Island was taken too")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("opponent lost %d, want 4", before-opp.Life)
	}
	if !me.Graveyard.Contains(blast) {
		t.Errorf("Fireblast cast from hand did not reach the graveyard — only flashback exiles")
	}
}

// Demon of Death's Gate: 6 life and three black creatures, both halves
// validated before either is paid.
func TestDemonOfDeathsGatePaysLifeAndThreeBlackCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	demon := effectstest.HandCardAtMain(t, g, "Demon of Death's Gate", "Creature — Demon", demonOfDeathsGateOracle)
	black := []uuid.UUID{
		effectstest.SeedColoredCreature(g, me.ID, "Zombie A", []string{"B"}, 0),
		effectstest.SeedColoredCreature(g, me.ID, "Zombie B", []string{"B"}, 0),
	}
	green := effectstest.SeedColoredCreature(g, me.ID, "Elf", []string{"G"}, 0)
	life := me.Life

	if err := effectstest.CastForAltCost(g, demon, "", "sacrifice", append(append([]uuid.UUID(nil), black...), green), nil); err == nil {
		t.Fatalf("a green creature paid \"three black creatures\"")
	}
	if me.Life != life {
		t.Errorf("the refused cast took %d life anyway", life-me.Life)
	}
	for _, id := range append(append([]uuid.UUID(nil), black...), green) {
		if !g.Battlefield.Contains(id) {
			t.Errorf("the refused cast sacrificed a creature anyway")
		}
	}

	black = append(black, effectstest.SeedColoredCreature(g, me.ID, "Zombie C", []string{"B"}, 0))
	if err := effectstest.CastForAltCost(g, demon, "", "sacrifice", black, nil); err != nil {
		t.Fatalf("Demon for 6 life and three black creatures: %v", err)
	}
	if me.Life != life-6 {
		t.Errorf("life = %d, want %d", me.Life, life-6)
	}
	for _, id := range black {
		if g.Battlefield.Contains(id) {
			t.Errorf("a black creature survived the cost")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(demon) {
		t.Fatalf("the Demon did not resolve onto the battlefield")
	}
	abilities := effectstest.EffectiveAbilities(t, g, demon)
	for _, kw := range []string{"flying", "trample"} {
		if !effectstest.ContainsString(abilities, kw) {
			t.Errorf("the Demon lacks %s: %v", kw, abilities)
		}
	}
}

// …and at 5 life the offer cannot be paid at all: refused, with all
// three creatures still alive.
func TestDemonOfDeathsGateRefusedBelowSixLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	demon := effectstest.HandCardAtMain(t, g, "Demon of Death's Gate", "Creature — Demon", demonOfDeathsGateOracle)
	black := []uuid.UUID{
		effectstest.SeedColoredCreature(g, me.ID, "Zombie A", []string{"B"}, 0),
		effectstest.SeedColoredCreature(g, me.ID, "Zombie B", []string{"B"}, 0),
		effectstest.SeedColoredCreature(g, me.ID, "Zombie C", []string{"B"}, 0),
	}
	me.Life = 5

	if err := effectstest.CastForAltCost(g, demon, "", "sacrifice", black, nil); err == nil {
		t.Fatalf("a player at 5 life paid 6")
	}
	for _, id := range black {
		if !g.Battlefield.Contains(id) {
			t.Errorf("the refused cast sacrificed a creature anyway")
		}
	}
	if me.Life != 5 || !me.Hand.Contains(demon) {
		t.Errorf("the refused cast changed life (%d) or moved the Demon", me.Life)
	}
}

// Register's boot checks on the component (#1727): a fixed count of at
// least one, over permanents, and one card-shaped payment per offer.
// Each refused shape compiles and would cast as a cheaper card than
// the one printed.
func TestRegisterRefusesAMalformedSacrificeAlternativeCost(t *testing.T) {
	withSpec := func(edit func(*game.AlternativeCost)) game.AlternativeCost {
		ac := effects.SacrificeInstead(2, "two creatures", 0, effects.Creature())
		edit(&ac)
		return ac
	}
	cases := []struct {
		name string
		ac   game.AlternativeCost
		want string
	}{
		{"sacrifice X", withSpec(func(ac *game.AlternativeCost) { ac.Sacrifice.CountFromX = true }),
			"no X to announce"},
		{"any number", withSpec(func(ac *game.AlternativeCost) { ac.Sacrifice.Min, ac.Sacrifice.Max = 0, 0 }),
			"any number"},
		{"one or more", withSpec(func(ac *game.AlternativeCost) { ac.Sacrifice.Min, ac.Sacrifice.Max = 1, 0 }),
			"variable count has no shape here"},
		{"a sacrifice and a pitch", withSpec(func(ac *game.AlternativeCost) {
			ac.ExileFromHand = effects.CardInYourHand("a card")
		}), "card-shaped payments"},
	}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := effectstest.RegisterPanics(effects.Spec{
				OracleID:         fmt.Sprintf("test-1727-register-%d", i),
				Name:             "Malformed Sacrifice Fixture",
				AlternativeCosts: []game.AlternativeCost{tc.ac},
			})
			if !strings.Contains(msg, tc.want) {
				t.Errorf("Register panic = %q, want one mentioning %q", msg, tc.want)
			}
		})
	}
}

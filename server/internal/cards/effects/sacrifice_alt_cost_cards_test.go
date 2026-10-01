package effects

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

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

// handCardAtMain seeds a card into the active seat's hand at a main
// phase, for a cast that claims an alternative cost.
func handCardAtMain(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id
}

func castForAltCost(g *game.Game, id uuid.UUID, from, key string, pay []uuid.UUID, targets []game.TargetRef) error {
	return g.CastSpell(g.Seats[g.Turn.ActiveSeat].ID, id, game.CastSpellParams{
		FromZone:        from,
		AlternativeCost: key,
		AltCostIDs:      pay,
		Targets:         targets,
	})
}

// Lava Dart: deals 1 from hand, then again from the graveyard for a
// Mountain — and the graveyard cast exiles it. An Island is not a
// Mountain.
func TestLavaDartFlashesBackForAMountain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	at := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	before := opp.Life

	dart := castCatalogSpell(t, g, "Lava Dart", "Instant", lavaDartOracle, at)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(dart) {
		t.Fatalf("Lava Dart cast from hand did not reach the graveyard")
	}

	island := seedLand(g, me.ID, "Island", "Basic Land — Island", "")
	if err := castForAltCost(g, dart, "graveyard", "flashback", []uuid.UUID{island}, at); err == nil {
		t.Fatalf("an Island paid \"Sacrifice a Mountain\"")
	}
	if !g.Battlefield.Contains(island) || !me.Graveyard.Contains(dart) {
		t.Fatalf("the refused flashback paid something anyway")
	}

	mountain := seedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	if err := castForAltCost(g, dart, "graveyard", "flashback", []uuid.UUID{mountain}, at); err != nil {
		t.Fatalf("flashback for a Mountain: %v", err)
	}
	if g.Battlefield.Contains(mountain) {
		t.Errorf("the Mountain survived the flashback cost")
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	at := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	before := opp.Life
	blast := handCardAtMain(t, g, "Fireblast", "Instant", fireblastOracle)
	m1 := seedLand(g, me.ID, "Mountain A", "Basic Land — Mountain", "")
	island := seedLand(g, me.ID, "Island", "Basic Land — Island", "")

	if err := castForAltCost(g, blast, "", "sacrifice", []uuid.UUID{m1, island}, at); err == nil {
		t.Fatalf("a Mountain and an Island paid \"sacrifice two Mountains\"")
	}
	if !g.Battlefield.Contains(m1) || !g.Battlefield.Contains(island) || !me.Hand.Contains(blast) {
		t.Fatalf("the refused cast paid something anyway")
	}

	m2 := seedLand(g, me.ID, "Mountain B", "Basic Land — Mountain", "")
	if err := castForAltCost(g, blast, "", "sacrifice", []uuid.UUID{m1, m2}, at); err != nil {
		t.Fatalf("Fireblast for two Mountains: %v", err)
	}
	if g.Battlefield.Contains(m1) || g.Battlefield.Contains(m2) {
		t.Errorf("a Mountain survived the cost")
	}
	if !g.Battlefield.Contains(island) {
		t.Errorf("the Island was taken too")
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	demon := handCardAtMain(t, g, "Demon of Death's Gate", "Creature — Demon", demonOfDeathsGateOracle)
	black := []uuid.UUID{
		seedColoredCreature(g, me.ID, "Zombie A", []string{"B"}, 0),
		seedColoredCreature(g, me.ID, "Zombie B", []string{"B"}, 0),
	}
	green := seedColoredCreature(g, me.ID, "Elf", []string{"G"}, 0)
	life := me.Life

	if err := castForAltCost(g, demon, "", "sacrifice", append(append([]uuid.UUID(nil), black...), green), nil); err == nil {
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

	black = append(black, seedColoredCreature(g, me.ID, "Zombie C", []string{"B"}, 0))
	if err := castForAltCost(g, demon, "", "sacrifice", black, nil); err != nil {
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
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(demon) {
		t.Fatalf("the Demon did not resolve onto the battlefield")
	}
	abilities := effectiveAbilities(t, g, demon)
	for _, kw := range []string{"flying", "trample"} {
		if !containsString(abilities, kw) {
			t.Errorf("the Demon lacks %s: %v", kw, abilities)
		}
	}
}

// …and at 5 life the offer cannot be paid at all: refused, with all
// three creatures still alive.
func TestDemonOfDeathsGateRefusedBelowSixLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	demon := handCardAtMain(t, g, "Demon of Death's Gate", "Creature — Demon", demonOfDeathsGateOracle)
	black := []uuid.UUID{
		seedColoredCreature(g, me.ID, "Zombie A", []string{"B"}, 0),
		seedColoredCreature(g, me.ID, "Zombie B", []string{"B"}, 0),
		seedColoredCreature(g, me.ID, "Zombie C", []string{"B"}, 0),
	}
	me.Life = 5

	if err := castForAltCost(g, demon, "", "sacrifice", black, nil); err == nil {
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
		ac := SacrificeInstead(2, "two creatures", 0, Creature())
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
			ac.ExileFromHand = CardInYourHand("a card")
		}), "card-shaped payments"},
	}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := registerPanics(Spec{
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

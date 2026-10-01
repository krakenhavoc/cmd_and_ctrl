package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// defender_controls_cards_test.go — the cards ADR 0107 §2 (#1879) lands:
// "This creature can't attack unless defending player controls <X>". The
// engine half (per-target defending player, planeswalkers, battles, the
// CR 508.1d interplay) is pinned in game/attack_unless_defender_controls_test.go.

// islandRestrictedOracles are the cards whose only rule beyond their
// keywords and body is "can't attack unless defending player controls an
// Island".
var islandRestrictedOracles = map[string]string{
	"Armored Galleon":            "637a10e8-4384-49a0-ad78-03da8930811e",
	"Deep-Sea Serpent":           "7fbb98cc-585c-4184-97f5-9b3d3ebdb1e5",
	"Ethereal Whiskergill":       "3d3e6dc0-2aed-4afa-bfa9-59d04ade9bee",
	"Hammerhead Shark":           "ce176172-6c7b-40b0-a6d0-68e9c32b3402",
	"Red Cliffs Armada":          "063884e0-1f5e-4be9-930b-e73895b2fa41",
	"Sea Monster":                "21b07ce7-b4f9-438c-8c09-e624557d62d2",
	"Sealock Monster":            "351c4f85-8792-4710-8cc9-d3e36657f6db",
	"Slipstream Eel":             "74e6dd0f-2866-4d45-a214-8b09b837bc02",
	"Steam Frigate":              "2e00598c-4f06-45aa-87a7-c63b5e8e92f3",
	"Vodalian Serpent":           "c39c1604-3bae-454d-9985-85101e51ec6e",
	"Wu Warship":                 "f184e860-05c3-43cf-a625-ab53427406c5",
	"Zhou Yu, Chief Commander":   "0b4742b7-e769-4354-beaf-6b4d18768ec1",
	"Serpent of the Endless Sea": "3b954d5f-3a93-4dd9-9d60-6097594d449c",
}

const (
	godhunterOctopusOracle   = "930b48b8-dbd1-4109-9eaa-7d7e9a04fa5b"
	lurkingGreenDragonOracle = "262e1cf6-61c8-444d-a71b-050eebcaf932"
	whimwaderOracle          = "21d9ce2c-eb6a-4f43-a79b-0b99b3dc4a00"
)

func pushTestPermanent(g *game.Game, controller uuid.UUID, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = controller, controller
	return pushBattlefieldCardWithTimestamp(g, c)
}

// declareResult tries one attack on a clone, so each check starts fresh.
func declareResult(g *game.Game, attacker, target uuid.UUID) error {
	return g.Clone().DeclareAttacker(attacker, target)
}

// TestIslandRestrictedCardsAttackOnlyOpponentsWithAnIsland — every
// Island card, in four seats: B controls an Island, C a Mountain, D
// nothing. Only B may be attacked. (The Serpent of the Endless Sea gets
// its own Island so it is not a 0/0.)
func TestIslandRestrictedCardsAttackOnlyOpponentsWithAnIsland(t *testing.T) {
	for name, oracle := range islandRestrictedOracles {
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, b, c, d := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
			pushTestPermanent(g, me.ID, game.Card{Name: "My Island", TypeLine: "Basic Land — Island"})
			pushTestPermanent(g, b.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
			pushTestPermanent(g, c.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
			id := pushTestPermanent(g, me.ID, game.Card{Name: name, OracleID: oracle, TypeLine: "Creature — Serpent", Power: 5, Toughness: 5})
			advanceToDeclareAttackersOf(t, g, 0)

			if err := declareResult(g, id, b.ID); err != nil {
				t.Errorf("attack the opponent with an Island: %v", err)
			}
			for _, refused := range []uuid.UUID{c.ID, d.ID} {
				if err := declareResult(g, id, refused); !errors.Is(err, game.ErrIllegalAttackTarget) {
					t.Errorf("attack an opponent with no Island: err = %v, want ErrIllegalAttackTarget", err)
				}
			}
		})
	}
}

// TestGodhunterOctopusWantsAnEnchantmentOrAnEnchantedPermanent — either
// half of the printed "or" lets it attack.
func TestGodhunterOctopusWantsAnEnchantmentOrAnEnchantedPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, b, c, d := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	octopus := pushTestPermanent(g, me.ID, game.Card{Name: "Godhunter Octopus", OracleID: godhunterOctopusOracle,
		TypeLine: "Creature — Octopus", Power: 5, Toughness: 5})
	pushTestPermanent(g, b.ID, game.Card{Name: "Test Enchantment", TypeLine: "Enchantment"})
	bear := pushTestPermanent(g, c.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	pushTestPermanent(g, me.ID, game.Card{Name: "Test Aura", TypeLine: "Enchantment — Aura",
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear}})
	pushTestPermanent(g, d.ID, game.Card{Name: "Wolf", TypeLine: "Creature — Wolf", Power: 2, Toughness: 2})
	advanceToDeclareAttackersOf(t, g, 0)

	for _, ok := range []uuid.UUID{b.ID, c.ID} {
		if err := declareResult(g, octopus, ok); err != nil {
			t.Errorf("attack %s: %v", ok, err)
		}
	}
	if err := declareResult(g, octopus, d.ID); !errors.Is(err, game.ErrIllegalAttackTarget) {
		t.Errorf("attack the player with neither: err = %v", err)
	}
}

// TestLurkingGreenDragonAndWhimwaderReadTheDefendersPermanents — "a
// creature with flying" and "a blue permanent", by current
// characteristics.
func TestLurkingGreenDragonAndWhimwaderReadTheDefendersPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	dragon := pushTestPermanent(g, me.ID, game.Card{Name: "Lurking Green Dragon", OracleID: lurkingGreenDragonOracle,
		TypeLine: "Creature — Dragon", Power: 4, Toughness: 4})
	wader := pushTestPermanent(g, me.ID, game.Card{Name: "Whimwader", OracleID: whimwaderOracle,
		TypeLine: "Creature — Elemental", Power: 6, Toughness: 4})
	// B: a blue creature with flying. C: a green creature without.
	pushTestPermanent(g, b.ID, game.Card{Name: "Blue Bird", TypeLine: "Creature — Bird", ManaCost: "{U}",
		Keywords: []string{"flying"}, Power: 1, Toughness: 1})
	pushTestPermanent(g, c.ID, game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", ManaCost: "{G}", Power: 2, Toughness: 2})
	advanceToDeclareAttackersOf(t, g, 0)

	for _, id := range []uuid.UUID{dragon, wader} {
		if err := declareResult(g, id, b.ID); err != nil {
			t.Errorf("attack B with %s: %v", id, err)
		}
		if err := declareResult(g, id, c.ID); !errors.Is(err, game.ErrIllegalAttackTarget) {
			t.Errorf("attack C with %s: err = %v", id, err)
		}
	}
	if !hasString(effectiveAbilities(t, g, dragon), "flying") {
		t.Error("Lurking Green Dragon has no flying")
	}
}

// TestSealockMonsterMakesItsOwnIsland — monstrous, it turns an opponent's
// Mountain into an Island in addition to its other types, and may then
// attack that opponent.
func TestSealockMonsterMakesItsOwnIsland(t *testing.T) {
	g := newCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	monster := pushTestPermanent(g, me.ID, game.Card{Name: "Sealock Monster", OracleID: islandRestrictedOracles["Sealock Monster"],
		TypeLine: "Creature — Octopus", Power: 5, Toughness: 5})
	mountain := pushTestPermanent(g, b.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})

	if err := g.ActivateCatalogAbility(me.ID, monster, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate monstrosity: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt for the becomes-monstrous trigger")
	}
	pickCard(t, g, me.ID, mountain)
	passPriorityAroundTable(t, g)

	land := findBattlefieldCardForTest(g, mountain)
	if land == nil || !land.HasSubtype("Island") || !land.HasSubtype("Mountain") {
		t.Fatalf("the Mountain is not an Island in addition to its other types: %+v", land)
	}
	if n := plusOneCounters(g, monster); n != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", n)
	}
	advanceToDeclareAttackersOf(t, g, 0)
	if err := declareResult(g, monster, b.ID); err != nil {
		t.Errorf("attack the opponent whose Mountain is now an Island: %v", err)
	}
}

// TestSerpentOfTheEndlessSeaCountsItsControllersIslands — the CDA counts
// the controller's Islands, not anyone else's.
func TestSerpentOfTheEndlessSeaCountsItsControllersIslands(t *testing.T) {
	g := newCatalogGame(t)
	me, b := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		pushTestPermanent(g, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	}
	pushTestPermanent(g, me.ID, game.Card{Name: "Volcanic Island", TypeLine: "Land — Island Mountain"})
	pushTestPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	pushTestPermanent(g, b.ID, game.Card{Name: "Their Island", TypeLine: "Basic Land — Island"})
	serpent := pushTestPermanent(g, me.ID, game.Card{Name: "Serpent of the Endless Sea",
		OracleID: islandRestrictedOracles["Serpent of the Endless Sea"], TypeLine: "Creature — Serpent"})
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	eff := findBattlefieldCardForTest(g, serpent).Effective()
	if eff.Power != 4 || eff.Toughness != 4 {
		t.Errorf("P/T = %d/%d, want 4/4 (four Islands, one of them nonbasic)", eff.Power, eff.Toughness)
	}
}

// TestVodalianSerpentEntersWithFourCountersOnlyWhenKicked — CR 702.33d's
// "if this creature was kicked", read off the cast.
func TestVodalianSerpentEntersWithFourCountersOnlyWhenKicked(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		want     int
	}{
		{"unkicked", nil, 0},
		{"kicked", []int{0}, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id, err := castWithOptionalCosts(t, g, "Vodalian Serpent", "Creature — Serpent",
				islandRestrictedOracles["Vodalian Serpent"], nil, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			if findBattlefieldCardForTest(g, id) == nil {
				t.Fatal("the Serpent did not enter")
			}
			if n := plusOneCounters(g, id); n != tc.want {
				t.Errorf("+1/+1 counters = %d, want %d", n, tc.want)
			}
		})
	}
}

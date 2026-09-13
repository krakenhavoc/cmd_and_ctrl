package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cost_modifier_test.go — S28 sub-PR 2: the nine cards that modify
// what other spells cost. The engine's own ordering and floor rules
// are pinned in game/cost_modifier_test.go; these tests are about
// the DECLARATIONS — that each card taxes or discounts the right
// spells cast by the right player, which is the half a card file can
// get wrong on its own.

const (
	sphereOfResistanceOracle = "09c96077-3804-4f12-a613-5bebc5e0413f"
	thornOfAmethystOracle    = "0c6c5336-9233-4ab1-9d55-79f20be7ea57"
	thaliaGuardianOracle     = "9b7f1d05-707c-4ed3-9f0e-8ced1232c2ee"
	goblinElectromancerOrcl  = "81f06f84-1580-43c0-89d5-08d34541a519"
	heartlessSummoningOracle = "fa1be67e-a07e-42ce-88be-446acd643dc6"
	auraOfSilenceOracle      = "e7faf8eb-e829-4109-8dfe-42865a23ba86"
	dampingSphereOracle      = "fd50d76c-7654-47b9-a5b6-d075874e4357"
	trinisphereOracle        = "69d994f2-b8f6-425f-9655-977c2144d40c"
	animarOracle             = "725880b2-1675-414f-b61b-cf6533797dbf"
)

// priceInHand seeds a spell into `seat`'s hand and returns what the
// board currently charges to cast it, in total mana value.
//
// Routed through the public ApplyCostModifiers surface rather than
// through a real cast, so a test can ask the question without also
// having to satisfy timing, priority and the mana pool.
func priceInHand(t *testing.T, g *game.Game, seat *game.Player, name, typeLine, manaCost string) int {
	t.Helper()
	base, err := game.ParseCost(manaCost)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", manaCost, err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{
		Card: game.Card{
			InstanceID: uuid.New(),
			Name:       name,
			TypeLine:   typeLine,
			ManaCost:   manaCost,
			Owner:      seat.ID,
			Controller: seat.ID,
		},
		Controller: seat.ID,
		FromZone:   game.ZoneHand,
	})
	if err != nil {
		t.Fatalf("ApplyCostModifiers: %v", err)
	}
	return out.ManaValue()
}

// Every card in the group has to actually reach the engine — a spec
// that declares modifiers the wire hook never returns is silently
// inert, which is the failure this catches.
func TestCostModifiersAreWired(t *testing.T) {
	for _, oracle := range []string{
		sphereOfResistanceOracle, thornOfAmethystOracle, thaliaGuardianOracle,
		goblinElectromancerOrcl, heartlessSummoningOracle, auraOfSilenceOracle,
		dampingSphereOracle, trinisphereOracle, animarOracle,
	} {
		if len(game.CostModifiersFor(oracle)) == 0 {
			t.Errorf("%s: no cost modifiers reached the engine", oracle)
		}
	}
	// The overwhelming majority of cards declare none, and the hook
	// has to say so rather than allocating an empty slice.
	if game.CostModifiersFor(faithlessLootingOracle) != nil {
		t.Errorf("Faithless Looting modifies nobody's costs")
	}
}

// Sphere of Resistance is symmetrical: it taxes its own controller
// exactly as it taxes everybody else.
func TestSphereOfResistanceTaxesEveryone(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Sphere of Resistance", "Artifact", sphereOfResistanceOracle, false)

	if got := priceInHand(t, g, me, "My Bolt", "Instant", "{R}"); got != 2 {
		t.Errorf("own spell under the sphere: %d, want 2", got)
	}
	if got := priceInHand(t, g, them, "Their Bear", "Creature — Bear", "{1}{G}"); got != 3 {
		t.Errorf("opponent's spell under the sphere: %d, want 3", got)
	}
}

// Thalia and Thorn of Amethyst share one clause; both must leave
// creature spells alone.
func TestNoncreatureTaxesSkipCreatures(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Thalia, Guardian of Thraben", "Legendary Creature — Human Soldier", thaliaGuardianOracle},
		{"Thorn of Amethyst", "Artifact", thornOfAmethystOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)

			if got := priceInHand(t, g, me, "Swords to Plowshares", "Instant", "{W}"); got != 2 {
				t.Errorf("noncreature spell: %d, want 2", got)
			}
			if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
				t.Errorf("creature spell: %d, want 2 (untaxed)", got)
			}
		})
	}
}

// Goblin Electromancer discounts only its controller's instants and
// sorceries — and only the generic half of them.
func TestGoblinElectromancerDiscountsYourInstantsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Goblin Electromancer", "Creature — Goblin Wizard", goblinElectromancerOrcl, false)

	if got := priceInHand(t, g, me, "Cultivate", "Sorcery", "{2}{G}"); got != 2 {
		t.Errorf("own sorcery: %d, want 2", got)
	}
	if got := priceInHand(t, g, them, "Their Cultivate", "Sorcery", "{2}{G}"); got != 3 {
		t.Errorf("opponent's sorcery: %d, want 3 (undiscounted)", got)
	}
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("own creature: %d, want 2 (undiscounted)", got)
	}
	// CR 601.2f: the reduction has no generic to spend here, and it
	// must not reach the {U}{U}.
	if got := priceInHand(t, g, me, "Counterspell", "Instant", "{U}{U}"); got != 2 {
		t.Errorf("{U}{U} under a {1} reduction: %d, want 2 (colours untouched)", got)
	}
}

// Heartless Summoning's headline gap: {2} off a one-mana coloured
// creature is still one mana, because a reduction spends generic.
func TestHeartlessSummoningStopsAtTheColoredHalf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Heartless Summoning", "Enchantment", heartlessSummoningOracle, false)

	if got := priceInHand(t, g, me, "Carrion Feeder", "Creature — Zombie", "{B}"); got != 1 {
		t.Errorf("{B} creature: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Solemn Simulacrum", "Artifact Creature — Golem", "{4}"); got != 2 {
		t.Errorf("{4} creature: %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Lightning Bolt", "Instant", "{R}"); got != 1 {
		t.Errorf("instant: %d, want 1 (creatures only)", got)
	}
}

// Aura of Silence is the asymmetric one — the predicate that would
// be easiest to omit, and the omission would look like it worked.
func TestAuraOfSilenceSparesItsOwnController(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Aura of Silence", "Enchantment", auraOfSilenceOracle, false)

	if got := priceInHand(t, g, me, "Sol Ring", "Artifact", "{1}"); got != 1 {
		t.Errorf("own artifact: %d, want 1 (untaxed)", got)
	}
	if got := priceInHand(t, g, them, "Sol Ring", "Artifact", "{1}"); got != 3 {
		t.Errorf("opponent's artifact: %d, want 3", got)
	}
	if got := priceInHand(t, g, them, "Rhystic Study", "Enchantment", "{2}{U}"); got != 5 {
		t.Errorf("opponent's enchantment: %d, want 5", got)
	}
	if got := priceInHand(t, g, them, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("opponent's creature: %d, want 2 (untaxed)", got)
	}
}

// Trinisphere raises to three in GENERIC mana and switches off when
// tapped.
func TestTrinisphereFloorsAtThreeAndRespectsTapping(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Trinisphere", "Artifact", trinisphereOracle, false)

	if got := priceInHand(t, g, me, "Dark Ritual", "Instant", "{B}"); got != 3 {
		t.Errorf("{B} under Trinisphere: %d, want 3", got)
	}
	if got := priceInHand(t, g, me, "Cultivate", "Sorcery", "{2}{G}"); got != 3 {
		t.Errorf("{2}{G} under Trinisphere: %d, want 3 (untouched)", got)
	}
	// Tap it: the predicate reads the live board, so the floor lifts.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if got := priceInHand(t, g, me, "Dark Ritual", "Instant", "{B}"); got != 1 {
		t.Errorf("{B} under a TAPPED Trinisphere: %d, want 1", got)
	}
}

// Damping Sphere scales with the caster's own spell count, and the
// "OTHER" in "each other spell that player has cast this turn" comes
// out of the tally being bumped after the price is settled.
func TestDampingSphereScalesWithSpellsAlreadyCast(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Damping Sphere", "Artifact", dampingSphereOracle, false)

	// Nothing cast yet: the first spell of the turn is untaxed.
	if got := priceInHand(t, g, me, "Lightning Bolt", "Instant", "{R}"); got != 1 {
		t.Errorf("first spell of the turn: %d, want 1", got)
	}

	// Cast two real spells, then re-price. The tally is per player,
	// so the opponent's count is untouched by mine.
	castCatalogSpell(t, g, "Filler A", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Filler B", "Instant", "", nil)
	passPriorityAroundTable(t, g)

	if got := priceInHand(t, g, me, "Lightning Bolt", "Instant", "{R}"); got != 3 {
		t.Errorf("third spell of the turn: %d, want 3 ({R} + {2})", got)
	}
	if got := priceInHand(t, g, them, "Lightning Bolt", "Instant", "{R}"); got != 1 {
		t.Errorf("opponent's first spell: %d, want 1 (tallies are per player)", got)
	}
}

// Animar discounts by its counters, and only its controller's
// creature spells.
func TestAnimarDiscountsPerCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	id := pushCatalogPermanent(g, me.ID, "Animar, Soul of Elements", "Legendary Creature — Elemental", animarOracle, false)

	if got := priceInHand(t, g, me, "Solemn Simulacrum", "Artifact Creature — Golem", "{4}"); got != 4 {
		t.Errorf("no counters: %d, want 4", got)
	}
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(id, game.CounterPlusOne, 3)
	})
	if got := priceInHand(t, g, me, "Solemn Simulacrum", "Artifact Creature — Golem", "{4}"); got != 1 {
		t.Errorf("three counters: %d, want 1", got)
	}
	if got := priceInHand(t, g, them, "Solemn Simulacrum", "Artifact Creature — Golem", "{4}"); got != 4 {
		t.Errorf("opponent's creature: %d, want 4 (undiscounted)", got)
	}
	// The coloured half survives any number of counters — which is
	// why Animar decks play colourless creatures.
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{1}{G}{G}"); got != 2 {
		t.Errorf("{1}{G}{G} with three counters: %d, want 2 (the {G}{G} stands)", got)
	}
}

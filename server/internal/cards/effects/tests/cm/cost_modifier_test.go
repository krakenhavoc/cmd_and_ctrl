package testscm

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Every card in the group has to actually reach the engine — a spec
// that declares modifiers the wire hook never returns is silently
// inert, which is the failure this catches.
func TestCostModifiersAreWired(t *testing.T) {
	for _, oracle := range []string{
		effectstest.SphereOfResistanceOracle, effectstest.ThornOfAmethystOracle, effectstest.ThaliaGuardianOracle,
		effectstest.GoblinElectromancerOrcl, effectstest.HeartlessSummoningOracle, effectstest.AuraOfSilenceOracle,
		effectstest.DampingSphereOracle, effectstest.TrinisphereOracle, effectstest.AnimarOracle,
	} {
		if len(game.CostModifiersFor(oracle)) == 0 {
			t.Errorf("%s: no cost modifiers reached the engine", oracle)
		}
	}
	// The overwhelming majority of cards declare none, and the hook
	// has to say so rather than allocating an empty slice.
	if game.CostModifiersFor(effectstest.FaithlessLootingOracle) != nil {
		t.Errorf("Faithless Looting modifies nobody's costs")
	}
}

// Sphere of Resistance is symmetrical: it taxes its own controller
// exactly as it taxes everybody else.
func TestSphereOfResistanceTaxesEveryone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Sphere of Resistance", "Artifact", effectstest.SphereOfResistanceOracle, false)

	if got := effectstest.PriceInHand(t, g, me, "My Bolt", "Instant", "{R}"); got != 2 {
		t.Errorf("own spell under the sphere: %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Their Bear", "Creature — Bear", "{1}{G}"); got != 3 {
		t.Errorf("opponent's spell under the sphere: %d, want 3", got)
	}
}

// Thalia and Thorn of Amethyst share one clause; both must leave
// creature spells alone.
func TestNoncreatureTaxesSkipCreatures(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Thalia, Guardian of Thraben", "Legendary Creature — Human Soldier", effectstest.ThaliaGuardianOracle},
		{"Thorn of Amethyst", "Artifact", effectstest.ThornOfAmethystOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)

			if got := effectstest.PriceInHand(t, g, me, "Swords to Plowshares", "Instant", "{W}"); got != 2 {
				t.Errorf("noncreature spell: %d, want 2", got)
			}
			if got := effectstest.PriceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
				t.Errorf("creature spell: %d, want 2 (untaxed)", got)
			}
		})
	}
}

// Goblin Electromancer discounts only its controller's instants and
// sorceries — and only the generic half of them.
func TestGoblinElectromancerDiscountsYourInstantsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Goblin Electromancer", "Creature — Goblin Wizard", effectstest.GoblinElectromancerOrcl, false)

	if got := effectstest.PriceInHand(t, g, me, "Cultivate", "Sorcery", "{2}{G}"); got != 2 {
		t.Errorf("own sorcery: %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Their Cultivate", "Sorcery", "{2}{G}"); got != 3 {
		t.Errorf("opponent's sorcery: %d, want 3 (undiscounted)", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("own creature: %d, want 2 (undiscounted)", got)
	}
	// CR 601.2f: the reduction has no generic to spend here, and it
	// must not reach the {U}{U}.
	if got := effectstest.PriceInHand(t, g, me, "Counterspell", "Instant", "{U}{U}"); got != 2 {
		t.Errorf("{U}{U} under a {1} reduction: %d, want 2 (colours untouched)", got)
	}
}

// Heartless Summoning's headline gap: {2} off a one-mana coloured
// creature is still one mana, because a reduction spends generic.
func TestHeartlessSummoningStopsAtTheColoredHalf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Heartless Summoning", "Enchantment", effectstest.HeartlessSummoningOracle, false)

	if got := effectstest.PriceInHand(t, g, me, "Carrion Feeder", "Creature — Zombie", "{B}"); got != 1 {
		t.Errorf("{B} creature: %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Solemn Simulacrum", "Artifact Creature — Golem", "{4}"); got != 2 {
		t.Errorf("{4} creature: %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Lightning Bolt", "Instant", "{R}"); got != 1 {
		t.Errorf("instant: %d, want 1 (creatures only)", got)
	}
}

// Aura of Silence is the asymmetric one — the predicate that would
// be easiest to omit, and the omission would look like it worked.
func TestAuraOfSilenceSparesItsOwnController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Aura of Silence", "Enchantment", effectstest.AuraOfSilenceOracle, false)

	if got := effectstest.PriceInHand(t, g, me, "Sol Ring", "Artifact", "{1}"); got != 1 {
		t.Errorf("own artifact: %d, want 1 (untaxed)", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Sol Ring", "Artifact", "{1}"); got != 3 {
		t.Errorf("opponent's artifact: %d, want 3", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Rhystic Study", "Enchantment", "{2}{U}"); got != 5 {
		t.Errorf("opponent's enchantment: %d, want 5", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("opponent's creature: %d, want 2 (untaxed)", got)
	}
}

// Trinisphere raises to three in GENERIC mana and switches off when
// tapped.
func TestTrinisphereFloorsAtThreeAndRespectsTapping(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Trinisphere", "Artifact", effectstest.TrinisphereOracle, false)

	if got := effectstest.PriceInHand(t, g, me, "Dark Ritual", "Instant", "{B}"); got != 3 {
		t.Errorf("{B} under Trinisphere: %d, want 3", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Cultivate", "Sorcery", "{2}{G}"); got != 3 {
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
	if got := effectstest.PriceInHand(t, g, me, "Dark Ritual", "Instant", "{B}"); got != 1 {
		t.Errorf("{B} under a TAPPED Trinisphere: %d, want 1", got)
	}
}

// Damping Sphere scales with the caster's own spell count, and the
// "OTHER" in "each other spell that player has cast this turn" comes
// out of the tally being bumped after the price is settled.
func TestDampingSphereScalesWithSpellsAlreadyCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Damping Sphere", "Artifact", effectstest.DampingSphereOracle, false)

	// Nothing cast yet: the first spell of the turn is untaxed.
	if got := effectstest.PriceInHand(t, g, me, "Lightning Bolt", "Instant", "{R}"); got != 1 {
		t.Errorf("first spell of the turn: %d, want 1", got)
	}

	// Cast two real spells, then re-price. The tally is per player,
	// so the opponent's count is untouched by mine.
	effectstest.CastCatalogSpell(t, g, "Filler A", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Filler B", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.PriceInHand(t, g, me, "Lightning Bolt", "Instant", "{R}"); got != 3 {
		t.Errorf("third spell of the turn: %d, want 3 ({R} + {2})", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Lightning Bolt", "Instant", "{R}"); got != 1 {
		t.Errorf("opponent's first spell: %d, want 1 (tallies are per player)", got)
	}
}

// Animar discounts by its counters, and only its controller's
// creature spells.
func TestAnimarDiscountsPerCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Animar, Soul of Elements", "Legendary Creature — Elemental", effectstest.AnimarOracle, false)

	if got := effectstest.PriceInHand(t, g, me, "Solemn Simulacrum", "Artifact Creature — Golem", "{4}"); got != 4 {
		t.Errorf("no counters: %d, want 4", got)
	}
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(id, game.CounterPlusOne, 3)
	})
	if got := effectstest.PriceInHand(t, g, me, "Solemn Simulacrum", "Artifact Creature — Golem", "{4}"); got != 1 {
		t.Errorf("three counters: %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Solemn Simulacrum", "Artifact Creature — Golem", "{4}"); got != 4 {
		t.Errorf("opponent's creature: %d, want 4 (undiscounted)", got)
	}
	// The coloured half survives any number of counters — which is
	// why Animar decks play colourless creatures.
	if got := effectstest.PriceInHand(t, g, me, "Bear", "Creature — Bear", "{1}{G}{G}"); got != 2 {
		t.Errorf("{1}{G}{G} with three counters: %d, want 2 (the {G}{G} stands)", got)
	}
}

// SpellManaValueAtLeast has no caller in the catalog yet (#832): this
// pins its one behaviour directly so it does not sit untested until a
// card needs it. {X}{U} announced at X=3 has mana value 4 (CR 202.3e)
// and passes a floor of 4; at X=2 its mana value is 3 and the same
// floor fails.
func TestSpellManaValueAtLeastReadsTheAnnouncedX(t *testing.T) {
	pred := effects.SpellManaValueAtLeast(4)
	card := game.Card{ManaCost: "{X}{U}"}
	if !pred(game.CostQuery{Card: card, XValue: 3}) {
		t.Error("X=3 ({X}{U} = mana value 4): want SpellManaValueAtLeast(4) to pass")
	}
	if pred(game.CostQuery{Card: card, XValue: 2}) {
		t.Error("X=2 ({X}{U} = mana value 3): want SpellManaValueAtLeast(4) to fail")
	}
}

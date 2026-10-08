package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exalted_test.go — #2538, the ADR 0101 amendment of 2026-10-08.
// Exalted is a canonical keyword whose trigger the engine derives
// (game/exalted.go); these tests hold the catalog half: the Hierarchs
// migrated off the old constructor with no double trigger, a printed
// exalted with no card file, and Emissary of Soulfire's counters.

const emissaryOfSoulfireOracle = "bdacd7a4-3d14-45eb-85a6-33ee711b34af"

// TestNoCatalogRowIsAnExaltedConstructor — A4 step 3: the retired
// constructor stays retired. A catalog row that watches attacks and
// calls itself exalted would trigger beside the keyword's own trigger.
func TestNoCatalogRowIsAnExaltedConstructor(t *testing.T) {
	for _, s := range All() {
		for _, row := range game.CatalogTriggers(s.OracleID) {
			if !strings.HasPrefix(row.Key, "Exalted") {
				continue
			}
			for _, w := range row.Watches {
				if w == game.EventAttack {
					t.Errorf("%s has a catalog exalted row %q: declare PrintedKeywords exalted instead", s.Name, row.Key)
				}
			}
		}
	}
}

// pushExaltedSource puts a 0/1 on me's battlefield, its printed
// keywords stamped as the deck importer stamps them.
func pushExaltedSource(g *game.Game, controller uuid.UUID, name, oracle string, keywords ...string) uuid.UUID {
	return apaPush(g, controller, controller, game.Card{
		Name: name, OracleID: oracle, TypeLine: "Creature — Human Druid",
		Power: 0, Toughness: 1, Keywords: keywords,
	})
}

// Each Hierarch declares exalted and the importer stamps it too; the
// two merge to one instance, so a lone attack is one +1/+1.
func TestEachHierarchHasOneExaltedTrigger(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Noble Hierarch", nobleHierarchOracle},
		{"Ignoble Hierarch", ignobleHierarchOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spec, _ := Lookup(tc.oracle)
			if len(spec.Triggered) != 0 {
				t.Errorf("%s still carries %d catalog triggers", tc.name, len(spec.Triggered))
			}
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceToMain(t, g)
			src := pushExaltedSource(g, me.ID, tc.name, tc.oracle, game.KeywordExalted)
			if n := game.ExaltedCount(apaLive(g, src)); n != 1 {
				t.Fatalf("%s has %d exalted instances, want 1", tc.name, n)
			}
			bear := seedBear(g, me.ID)
			attackWith(t, g, opp.ID, bear)
			if got := effectivePower(t, g, bear); got != 3 {
				t.Errorf("power %d, want 3 (2 base + one exalted)", got)
			}
		})
	}
}

// Two exalted sources are two triggers; a printed exalted on a card
// with no catalog entry (Akrasan Squire) works with no card file.
func TestExaltedSourcesAddUp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushExaltedSource(g, me.ID, "Noble Hierarch", nobleHierarchOracle)
	pushExaltedSource(g, me.ID, "Akrasan Squire", "", game.KeywordExalted)
	bear := seedBear(g, me.ID)
	attackWith(t, g, opp.ID, bear)
	if got, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); got != 4 || tough != 4 {
		t.Errorf("bear is %d/%d, want 4/4 (two exalted instances)", got, tough)
	}
}

// A layer-6 grant (Sublime Archangel's "other creatures you control
// have exalted") gives each creature an instance of its own.
func TestGrantedExaltedTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	elf := pushVanillaCreature(g, me.ID, "Elf", 1, 1)
	bear := seedBear(g, me.ID)
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(elf,
			[]game.AffectedObject{game.PinObject(elf, findBattlefieldCardForTest(g, elf).EnteredBattlefieldAt)},
			[]game.Mod{game.AddKeywordsMod(game.KeywordExalted)}, game.IndefiniteDuration(), "test grant")
	})
	attackWith(t, g, opp.ID, bear)
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3 (one granted exalted)", got)
	}
}

func TestEmissaryOfSoulfire(t *testing.T) {
	spec, ok := Lookup(emissaryOfSoulfireOracle)
	if !ok || spec.Completeness != CompletenessFull {
		t.Fatalf("Emissary of Soulfire is not a Full catalog card")
	}
	if c := specActivatedEnergy(t, emissaryOfSoulfireOracle, 0); c.Energy != 2 {
		t.Errorf("activation costs %d energy, want 2", c.Energy)
	}

	// Entering gives three energy.
	g, me, _ := p7Table(t)
	castCatalogSpell(t, g, "Emissary of Soulfire", "Creature — Djinn Monk", emissaryOfSoulfireOracle, nil)
	passPriorityAroundTable(t, g)
	if got := energyOf(me); got != 3 {
		t.Errorf("energy after Emissary entered = %d, want 3", got)
	}

	// Two activations put two exalted counters on a creature I control:
	// two instances, two triggers on a lone attack.
	g, me, opp := p7Table(t)
	src := ebPush(g, me, "Emissary of Soulfire", emissaryOfSoulfireOracle, "Creature — Djinn Monk", 1, 4)
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Targets: ebCardTarget(theirs)}); err == nil {
		t.Fatal("Emissary targeted a creature I don't control")
	}
	ebActivate(t, g, me, src, 0, 2, ebCardTarget(bear))
	ebActivate(t, g, me, src, 0, 2, ebCardTarget(bear))
	live := ebLive(t, g, bear)
	if live.Counters[game.CounterExalted] != 2 || game.ExaltedCount(live) != 2 {
		t.Fatalf("bear has %d exalted counters and %d instances, want 2 and 2",
			live.Counters[game.CounterExalted], game.ExaltedCount(live))
	}
	attackWith(t, g, opp.ID, bear)
	if got, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); got != 4 || tough != 4 {
		t.Errorf("bear is %d/%d after a lone attack, want 4/4 (two exalted counters)", got, tough)
	}

	// Activate only as a sorcery: not during combat.
	setEnergy(t, g, me, 2)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Targets: ebCardTarget(bear)}); err == nil {
		t.Error("Emissary activated outside a main phase")
	}
}

// A counter on another permanent counts too: exalted triggers from any
// permanent you control, not only the attacker (the reminder text,
// "for each instance of exalted among permanents you control").
func TestAnExaltedCounterOnANonAttackerCounts(t *testing.T) {
	g, me, opp := p7Table(t)
	src := ebPush(g, me, "Emissary of Soulfire", emissaryOfSoulfireOracle, "Creature — Djinn Monk", 1, 4)
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	ebActivate(t, g, me, src, 0, 2, ebCardTarget(src))
	attackWith(t, g, opp.ID, bear)
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3 (the Emissary's exalted counter)", got)
	}
}

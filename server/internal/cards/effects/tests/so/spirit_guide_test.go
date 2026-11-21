package testsso

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The constructor builds exactly the printed clause: the hand zone,
// the exile-this cost, no tap and no sacrifice.
func TestExileFromHandForManaDeclaresZoneAndCost(t *testing.T) {
	ab := effects.ExileFromHandForMana("{R}")
	if ab.Produced != "{R}" {
		t.Errorf("Produced = %q, want {R}", ab.Produced)
	}
	if len(ab.Zones) != 1 || ab.Zones[0] != game.ZoneHand {
		t.Errorf("Zones = %v, want [hand]", ab.Zones)
	}
	if !ab.Cost.ExileSelf {
		t.Error("the cost does not exile the card")
	}
	if ab.Cost.Tap || ab.Cost.Sacrifice || ab.Cost.Mana != "" || ab.Cost.Life != 0 {
		t.Errorf("cost = %+v, want the exile and nothing else", ab.Cost)
	}
	if !strings.Contains(ab.Label, "from your hand") {
		t.Errorf("label = %q, want the printed clause", ab.Label)
	}
}

// The shape reaches the ENGINE through the def, which is the thing
// every consumer actually reads. A Spec that declared the zone and
// lost it in buildDef would look right here and never work in play.
func TestSpiritGuideDefCarriesTheZoneAndTheCost(t *testing.T) {
	for _, name := range []string{"Simian Spirit Guide", "Elvish Spirit Guide"} {
		spec := effectstest.SpecByName(t, name)
		def := effects.BuildDef(spec)
		if len(def.ManaAbilities) != 1 {
			t.Fatalf("%s: %d mana abilities, want 1", name, len(def.ManaAbilities))
		}
		ab := def.ManaAbilities[0]
		if !game.ManaAbilityFunctionsFromZone(ab, game.ZoneHand) {
			t.Errorf("%s: the def's ability does not function from a hand", name)
		}
		if game.ManaAbilityFunctionsFromZone(ab, game.ZoneBattlefield) {
			t.Errorf("%s: the def's ability also functions from the battlefield", name)
		}
		if !ab.ExileSelf {
			t.Errorf("%s: the def lost the exile-this cost", name)
		}
	}
}

// The two cards produce the two colours they print, which is the
// whole of what distinguishes them — and therefore the whole of what
// a shared constructor could get wrong.
func TestSpiritGuidesProduceTheirPrintedColours(t *testing.T) {
	cases := map[string]string{
		"Simian Spirit Guide": "{R}",
		"Elvish Spirit Guide": "{G}",
	}
	for name, want := range cases {
		spec := effectstest.SpecByName(t, name)
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness = %v, want full — the whole card is the ability",
				name, spec.Completeness)
		}
		if len(spec.ManaAbilities) != 1 {
			t.Fatalf("%s: %d mana abilities, want 1", name, len(spec.ManaAbilities))
		}
		if got := spec.ManaAbilities[0].Produced; got != want {
			t.Errorf("%s produces %q, want %q", name, got, want)
		}
	}
}

// Boot refusal 1: a zone no consumer walks. The card would register,
// look complete on the catalog page, and never offer the ability.
func TestRegisterRefusesAnUnwalkedManaZone(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("an exile mana ability registered — nothing walks exile for one")
		}
	}()
	effects.Register(effects.Spec{
		OracleID: "test-mana-zone-exile",
		Name:     "Exiled Ritual",
		ManaAbilities: []effects.ManaAbility{{
			Zones:    []game.ZoneKind{game.ZoneExile},
			Produced: "{B}",
			Cost:     effects.ManaAbilityCost{ExileSelf: true},
		}},
	})
}

// Boot refusal 2: a component only a permanent could pay, on an
// ability that functions off the battlefield. There is nothing in a
// hand to tap.
func TestRegisterRefusesATapCostOnAHandManaAbility(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("a hand mana ability with a tap cost registered")
		}
	}()
	effects.Register(effects.Spec{
		OracleID: "test-mana-zone-tap",
		Name:     "Tapping Guide",
		ManaAbilities: []effects.ManaAbility{{
			Zones:    []game.ZoneKind{game.ZoneHand},
			Produced: "{R}",
			Cost:     effects.ManaAbilityCost{Tap: true, ExileSelf: true},
		}},
	})
}

// Boot refusal 3, both directions of the same implication. A hand
// zone with no exile cost is a free repeatable mana source; an exile
// cost with no hand zone has nothing to exile from.
func TestRegisterRefusesAHalfDeclaredHandManaAbility(t *testing.T) {
	t.Run("zone without cost", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("a costless hand mana ability registered")
			}
		}()
		effects.Register(effects.Spec{
			OracleID: "test-mana-zone-free",
			Name:     "Free Ritual",
			ManaAbilities: []effects.ManaAbility{{
				Zones:    []game.ZoneKind{game.ZoneHand},
				Produced: "{R}",
			}},
		})
	})
	t.Run("cost without zone", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("an exile-this mana ability with no zone registered")
			}
		}()
		effects.Register(effects.Spec{
			OracleID: "test-mana-zone-orphan-cost",
			Name:     "Orphan Ritual",
			ManaAbilities: []effects.ManaAbility{{
				Produced: "{R}",
				Cost:     effects.ManaAbilityCost{ExileSelf: true},
			}},
		})
	})
}

// --- the two proof cards, end to end -------------------------------

const (
	simianSpiritGuideOracle = "44e0ffa3-8915-4c1f-8f1a-4aeea1365f07"
	elvishSpiritGuideOracle = "6b0e23cf-7d68-4329-86db-7adc26abd86b"
)

// activateSpiritGuide seeds the named card into the active seat's hand
// and fires its mana ability, returning the card and the seat. The
// hand seeder is cycling_test.go's — the two keyword families
// activate out of the same pile and there is no reason for two.
func activateSpiritGuide(t *testing.T, g *game.Game, name, oracle string) (uuid.UUID, *game.Player) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushCatalogHandCard(me, name, "Creature — Ape Spirit", oracle)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("%s: activate from hand: %v", name, err)
	}
	return id, me
}

// Simian Spirit Guide: {R} in the pool, the card in exile, and
// nothing on the stack in between (CR 605.3b).
func TestSimianSpiritGuideAddsRedFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := activateSpiritGuide(t, g, "Simian Spirit Guide", simianSpiritGuideOracle)

	if me.Hand.Contains(id) {
		t.Error("the Spirit Guide is still in hand")
	}
	if !g.Exile.Contains(id) {
		t.Error("the Spirit Guide did not reach exile")
	}
	if len(g.Stack.Cards) != 0 {
		t.Error("a mana ability used the stack (CR 605.3b)")
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "R" {
		t.Errorf("mana pool = %v, want one {R}", me.ManaPool)
	}
}

// Elvish Spirit Guide: the same card, the other colour. The
// constructor reads its argument rather than the engine knowing the
// cards.
func TestElvishSpiritGuideAddsGreenFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := activateSpiritGuide(t, g, "Elvish Spirit Guide", elvishSpiritGuideOracle)

	if !g.Exile.Contains(id) {
		t.Error("the Spirit Guide did not reach exile")
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "G" {
		t.Errorf("mana pool = %v, want one {G}", me.ManaPool)
	}
}

// And the auto-tapper finds them through the catalog: a Spirit Guide
// in hand pays a {R} cast with nothing on the battlefield — the path
// the cast preview, the strict gate and every bot go through.
func TestSpiritGuideIsAnAutoTapSourceFromTheCatalog(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushCatalogHandCard(me, "Simian Spirit Guide", "Creature — Ape Spirit", simianSpiritGuideOracle)

	cost, err := game.ParseCost("{R}")
	if err != nil {
		t.Fatalf("ParseCost: %v", err)
	}
	plan, ok := g.AutoTapForCost(me.ID, cost, 0)
	if !ok || len(plan) != 1 || plan[0] != id {
		t.Fatalf("plan = %v ok = %v, want the Spirit Guide", plan, ok)
	}
}

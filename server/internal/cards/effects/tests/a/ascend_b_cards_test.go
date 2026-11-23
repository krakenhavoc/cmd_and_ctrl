package testsa

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ascend_b_cards_test.go — the remaining ascend cards (#2706) that
// need nothing the engine lacks. The vocabulary is citys_blessing.go.

const (
	snubhornSentryOracle        = "fc8d3933-f7fe-4849-aae7-737ce6f067de"
	duskChargerOracle           = "e0190e13-89f6-4b51-a277-bf3eb842963e"
	spireWinderOracle           = "e8cfb24a-7e52-4e39-a9e1-969b4a65b62e"
	skymarcherAspirantOracle    = "0b6db929-1b6a-4372-8146-aa047258b552"
	stormFleetSwashbucklerOra   = "6640bcfc-4bdc-4380-8b51-61b906f9aee6"
	resplendentGriffinOracle    = "f8bf5df2-b783-4796-b2c6-eb4dd4167d62"
	mausoleumHarpyOracle        = "ba682da8-50f5-478f-99cc-60a57d229a04"
	deadeyeBrawlerOracle        = "a7c36b66-4f58-41af-b1c5-704ff0d9eac8"
	prideOfConquerorsOracleText = "e45a1c99-a020-49ab-8971-35c6bcd096c8"
)

func TestAscendBCardsDeclareTheKeyword(t *testing.T) {
	for name, oracle := range map[string]string{
		"Snubhorn Sentry": snubhornSentryOracle, "Dusk Charger": duskChargerOracle,
		"Spire Winder": spireWinderOracle, "Skymarcher Aspirant": skymarcherAspirantOracle,
		"Storm Fleet Swashbuckler": stormFleetSwashbucklerOra, "Resplendent Griffin": resplendentGriffinOracle,
		"Mausoleum Harpy": mausoleumHarpyOracle, "Deadeye Brawler": deadeyeBrawlerOracle,
		"Pride of Conquerors": prideOfConquerorsOracleText,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if !effectstest.HasKeyword(spec.PrintedKeywords, game.KeywordAscend) {
			t.Errorf("%s does not declare ascend", name)
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s completeness = %v, want full", name, spec.Completeness)
		}
	}
}

func TestSelfPumpsOnlyWithTheBlessing(t *testing.T) {
	for _, tc := range []struct {
		name, typ, oracle string
		p, tough, dp, dt  int
	}{
		{"Snubhorn Sentry", "Creature — Dinosaur", snubhornSentryOracle, 0, 3, 3, 0},
		{"Dusk Charger", "Creature — Horse", duskChargerOracle, 3, 3, 2, 2},
		{"Spire Winder", "Creature — Snake", spireWinderOracle, 2, 3, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typ, tc.oracle, tc.p, tc.tough)
			other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			if effectstest.EffectivePower(t, g, id) != tc.p || effectstest.EffectiveToughness(t, g, id) != tc.tough {
				t.Fatalf("pumped without the blessing: %d/%d", effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id))
			}
			// An opponent's blessing is not yours.
			effectstest.GrantBlessing(g, opp)
			if effectstest.EffectivePower(t, g, id) != tc.p {
				t.Error("pumped by an opponent's blessing")
			}
			effectstest.GrantBlessing(g, me)
			if got := effectstest.EffectivePower(t, g, id); got != tc.p+tc.dp {
				t.Errorf("power %d, want %d", got, tc.p+tc.dp)
			}
			if got := effectstest.EffectiveToughness(t, g, id); got != tc.tough+tc.dt {
				t.Errorf("toughness %d, want %d", got, tc.tough+tc.dt)
			}
			if effectstest.EffectivePower(t, g, other) != 2 {
				t.Error("the pump leaked onto another creature")
			}
		})
	}
}

func TestSelfKeywordOnlyWithTheBlessing(t *testing.T) {
	for _, tc := range []struct {
		name, typ, oracle, kw string
	}{
		{"Skymarcher Aspirant", "Creature — Vampire Soldier", skymarcherAspirantOracle, "flying"},
		{"Storm Fleet Swashbuckler", "Creature — Human Pirate", stormFleetSwashbucklerOra, "double strike"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.B12Push(g, me.ID, tc.name, tc.typ, tc.oracle, 2, 2)
			other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			if effectstest.HasString(effectstest.EffectiveAbilities(t, g, id), tc.kw) {
				t.Fatalf("%s without the blessing", tc.kw)
			}
			effectstest.GrantBlessing(g, me)
			if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, id), tc.kw) {
				t.Errorf("no %s with the blessing", tc.kw)
			}
			if effectstest.HasString(effectstest.EffectiveAbilities(t, g, other), tc.kw) {
				t.Errorf("another creature got %s", tc.kw)
			}
		})
	}
}

func TestResplendentGriffinGrowsOnlyWithTheBlessing(t *testing.T) {
	for _, blessed := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		griffin := effectstest.B12Push(g, me.ID, "Resplendent Griffin", "Creature — Griffin", resplendentGriffinOracle, 2, 2)
		if blessed {
			effectstest.GrantBlessing(g, me)
		}
		effectstest.BrAttack(t, g, griffin)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if blessed {
			want = 1
		}
		if got := effectstest.PlusOneCounters(g, griffin); got != want {
			t.Errorf("blessed=%v: %d +1/+1 counters, want %d", blessed, got, want)
		}
	}
}

func TestMausoleumHarpyGrowsWhenAnotherOfYoursDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harpy := effectstest.B12Push(g, me.ID, "Mausoleum Harpy", "Creature — Harpy", mausoleumHarpyOracle, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	// Without the blessing nothing happens.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, harpy); got != 0 {
		t.Fatalf("grew without the blessing: %d", got)
	}

	effectstest.GrantBlessing(g, me)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, harpy); got != 0 {
		t.Errorf("grew for an opponent's creature: %d", got)
	}

	mine2 := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine2) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, harpy); got != 1 {
		t.Errorf("counters %d, want 1", got)
	}

	// "Another": the Harpy's own death grows nothing (and it is gone).
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(harpy) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(harpy) {
		t.Error("the Harpy survived its destruction")
	}
}

func TestDeadeyeBrawlerDrawsOnlyWithTheBlessingAndOnlyForPlayers(t *testing.T) {
	for _, blessed := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		brawler := effectstest.B12Push(g, me.ID, "Deadeye Brawler", "Creature — Human Pirate", deadeyeBrawlerOracle, 2, 4)
		if blessed {
			effectstest.GrantBlessing(g, me)
		}
		hand := me.Hand.Size()
		effectstest.DealCombatDamageToPlayer(g, brawler, opp.ID, 2)
		effectstest.PassPriorityAroundTable(t, g)
		want := hand
		if blessed {
			want = hand + 1
		}
		if me.Hand.Size() != want {
			t.Errorf("blessed=%v: hand %d -> %d, want %d", blessed, hand, me.Hand.Size(), want)
		}
	}
}

func TestDeadeyeBrawlerIgnoresOtherCreaturesDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Deadeye Brawler", "Creature — Human Pirate", deadeyeBrawlerOracle, 2, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.GrantBlessing(g, me)
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("drew for another creature's combat damage")
	}
}

func TestPrideOfConquerorsPumpsOneOrTwo(t *testing.T) {
	for _, tc := range []struct {
		name    string
		blessed bool
		want    int
	}{{"without the blessing", false, 3}, {"with the blessing", true, 4}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
			if tc.blessed {
				effectstest.GrantBlessing(g, me)
			}
			effectstest.CastCatalogSpell(t, g, "Pride of Conquerors", "Instant", prideOfConquerorsOracleText, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.EffectivePower(t, g, mine); got != tc.want {
				t.Errorf("power %d, want %d", got, tc.want)
			}
			if got := effectstest.EffectiveToughness(t, g, mine); got != tc.want {
				t.Errorf("toughness %d, want %d", got, tc.want)
			}
			if effectstest.EffectivePower(t, g, theirs) != 2 {
				t.Error("an opponent's creature was pumped")
			}
		})
	}
}

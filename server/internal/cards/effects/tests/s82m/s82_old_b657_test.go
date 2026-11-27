package testss82m

import (
	"reflect"
	"sort"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b657_test.go — Invasion, batch 657 (#5616): mana and ramp.

type oldB657Row struct{ oracle, name string }

var oldB657Rows = []oldB657Row{
	{"1ce6ae30-33c3-4f05-9286-69b0871b1c2d", "Bloodstone Cameo"},
	{"23f93411-f83c-4ed2-abed-99cf905f7d7f", "Bog Initiate"},
	{"383f6020-c26d-43e8-bb07-566886626d74", "Seashell Cameo"},
	{"6223b1c0-bfe0-490c-b2d4-28537b05f571", "Crosis's Attendant"},
	{"70bb2690-653b-4787-bdd1-d1739a8ae545", "Quirion Sentinel"},
	{"781542b7-155b-4a66-aa52-eec0ebb29bb2", "Rith's Attendant"},
	{"8fbdec25-4222-4b96-aeea-82a4b5b8b80e", "Drake-Skull Cameo"},
	{"91d36a9d-3a3f-4cb7-837d-b1e8f6730512", "Dromar's Attendant"},
	{"98e042de-05f4-4e2e-b12f-375b905e6600", "Troll-Horn Cameo"},
	{"c0ad16d7-2550-47e0-8581-0e22b27cb0d0", "Darigaaz's Attendant"},
	{"e2975a88-5a5a-462a-90b9-d8c00150b6d3", "Quirion Elves"},
	{"ff7d927f-0a01-4243-ad00-c65686ee86bb", "Treva's Attendant"},
}

func TestOldB657EveryCardIsRegisteredFull(t *testing.T) {
	for _, r := range oldB657Rows {
		spec, ok := effects.Lookup(r.oracle)
		if !ok {
			t.Errorf("%s is not registered", r.name)
			continue
		}
		if spec.Name != r.name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", r.name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB657CameosTapForEitherColour(t *testing.T) {
	for _, tc := range []struct{ oracle, name, a, b string }{
		{"1ce6ae30-33c3-4f05-9286-69b0871b1c2d", "Bloodstone Cameo", "B", "R"},
		{"383f6020-c26d-43e8-bb07-566886626d74", "Seashell Cameo", "W", "U"},
		{"8fbdec25-4222-4b96-aeea-82a4b5b8b80e", "Drake-Skull Cameo", "U", "B"},
		{"98e042de-05f4-4e2e-b12f-375b905e6600", "Troll-Horn Cameo", "R", "G"},
	} {
		for _, want := range []string{tc.a, tc.b} {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Artifact", tc.oracle, false)
			if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{want}}); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{want}) {
				t.Errorf("%s: pool %v, want [%s]", tc.name, got, want)
			}
			if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{want}}); err == nil {
				t.Errorf("%s: a tapped Cameo made mana again", tc.name)
			}
		}
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Artifact", tc.oracle, false)
		other := "W"
		if tc.a == "W" || tc.b == "W" {
			other = "G"
		}
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{other}}); err == nil {
			t.Errorf("%s: a colour it does not print was accepted", tc.name)
		}
	}
}

func TestOldB657AttendantsSacrificeForThreeColours(t *testing.T) {
	for _, tc := range []struct {
		oracle, name string
		want         []string
	}{
		{"6223b1c0-bfe0-490c-b2d4-28537b05f571", "Crosis's Attendant", []string{"B", "R", "U"}},
		{"781542b7-155b-4a66-aa52-eec0ebb29bb2", "Rith's Attendant", []string{"G", "R", "W"}},
		{"91d36a9d-3a3f-4cb7-837d-b1e8f6730512", "Dromar's Attendant", []string{"B", "U", "W"}},
		{"c0ad16d7-2550-47e0-8581-0e22b27cb0d0", "Darigaaz's Attendant", []string{"B", "G", "R"}},
		{"ff7d927f-0a01-4243-ad00-c65686ee86bb", "Treva's Attendant", []string{"G", "U", "W"}},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Artifact Creature — Golem", tc.oracle, true)
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err == nil {
			t.Fatalf("%s: activated with no {1} to pay", tc.name)
		}
		if !effectstest.OnBattlefieldByID(g, id) {
			t.Fatalf("%s: sacrificed without paying", tc.name)
		}
		effectstest.FloatMana(t, g, me, "{C}")
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: %v (a summoning-sick Attendant has no {T} in its cost)", tc.name, err)
		}
		if effectstest.OnBattlefieldByID(g, id) {
			t.Errorf("%s: not sacrificed", tc.name)
		}
		if got := oldB657Sorted(me); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: pool %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestOldB657BogInitiatePaysOneForBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Bog Initiate", "Creature — Human Wizard", "23f93411-f83c-4ed2-abed-99cf905f7d7f", true)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("activated with no {1} to pay")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("pool %v, want [B]", got)
	}
	// No {T} in the cost: it works again with another {1}.
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("second activation: %v", err)
	}
}

func TestOldB657QuirionSentinelAddsOneManaOfAnyColourWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.RfCrACast(t, g, "Quirion Sentinel", "Creature — Elf Druid", "70bb2690-653b-4787-bdd1-d1739a8ae545", 2, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick != nil {
		if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
			t.Fatalf("colour: %v", err)
		}
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool %v, want [R]", got)
	}
}

func TestOldB657QuirionElvesTapForGreenOrTheChosenColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elves := effectstest.CastCatalogSpell(t, g, "Quirion Elves", "Creature — Elf Druid", "e2975a88-5a5a-462a-90b9-d8c00150b6d3", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "U")
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, elves).SummonedThisTurn = false })
	if err := g.ActivateManaAbility(me.ID, elves, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("chosen-colour ability: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("pool %v, want [U]", got)
	}
	effectstest.Untap(g, elves)
	if err := g.ActivateManaAbility(me.ID, elves, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("green ability: %v", err)
	}
	if got := oldB657Sorted(me); !reflect.DeepEqual(got, []string{"G", "U"}) {
		t.Errorf("pool %v, want [G U]", got)
	}
}

// oldB657Sorted is the pool's colours in alphabetical order: the pool keeps
// the order the mana arrived in, which these tests do not assert.
func oldB657Sorted(p *game.Player) []string {
	got := append([]string(nil), effectstest.PoolColors(p)...)
	sort.Strings(got)
	return got
}

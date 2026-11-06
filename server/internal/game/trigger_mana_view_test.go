package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// trigger_mana_view_test.go is #4770: TapTriggerMana, the view's reader
// of what a tap adds through triggered mana abilities, answers from the
// planner's judge, on the planner tests' stubbed triggers.

// tapTriggerAdds is what TapTriggerMana says each of the card's mana rows
// would add for `who`, in row order; nil when the board has no triggered
// mana ability at all.
func tapTriggerAdds(t *testing.T, g *Game, who *Player, card uuid.UUID) []string {
	t.Helper()
	var out []string
	g.ReadSnapshot(func() {
		tm, ok := g.TapTriggerManaLocked()
		if !ok {
			return
		}
		c, found := g.LookupCardPtrForEffect(card)
		if !found {
			t.Fatalf("card %s is not on the battlefield", card)
		}
		rows := ManaAbilitiesOf(c)
		out = make([]string, len(rows))
		for i := range rows {
			out[i] = tm.AddsLocked(who.ID, c, &rows[i])
		}
	})
	return out
}

func TestTapTriggerManaIsAbsentWithoutATrigger(t *testing.T) {
	g, me, _ := triggerTable(t)
	forest := pushForest(g, me)
	if got := tapTriggerAdds(t, g, me, forest); got != nil {
		t.Errorf("no triggered mana ability on the board, yet the reader answered %v", got)
	}
}

func TestTapTriggerManaNamesAFixedTrigger(t *testing.T) {
	g, me, opp := triggerTable(t)
	pushBattlefieldForTest(g, me.ID, "Crypt Ghast", "Creature — Spirit", plannerGhastOracle)
	swamp := pushLandFor(g, me, "Swamp", "Basic Land — Swamp")
	island := pushLandFor(g, me, "Island", "Basic Land — Island")
	theirs := pushLandFor(g, opp, "Swamp", "Basic Land — Swamp")
	grave := pushPlannerSource(g, me, "Watery Grave", "Land — Island Swamp", "", tapAbility("{U|B}"))

	for _, tc := range []struct {
		name string
		who  *Player
		card uuid.UUID
		want string // what every row of the card adds
	}{
		{"a Swamp under Crypt Ghast", me, swamp, "{B}"},
		{"an Island under Crypt Ghast", me, island, ""},
		// "Whenever YOU tap a Swamp": the opponent's Swamp gets nothing.
		{"an opponent's Swamp", opp, theirs, ""},
		// The dual's extra {B} is the same whichever colour it makes, so
		// it is named, on its own row and its intrinsic ones alike.
		{"a Swamp-typed dual", me, grave, "{B}"},
	} {
		got := tapTriggerAdds(t, g, tc.who, tc.card)
		if len(got) == 0 {
			t.Errorf("%s: no mana rows", tc.name)
		}
		for i := range got {
			if got[i] != tc.want {
				t.Errorf("%s: row %d adds %q, want %q (all rows %v)", tc.name, i, got[i], tc.want, got)
			}
		}
	}
}

func TestTapTriggerManaOfAnAuraAndOfAnyColour(t *testing.T) {
	g, me, _ := triggerTable(t)
	grown := pushForest(g, me)
	pushAuraOn(g, me, "Wild Growth", plannerWildGrowthOracle, grown)
	fertile := pushForest(g, me)
	pushAuraOn(g, me, "Fertile Ground", plannerFertileOracle, fertile)
	plain := pushForest(g, me)

	if got := tapTriggerAdds(t, g, me, grown); !reflect.DeepEqual(got, []string{"{G}"}) {
		t.Errorf("Forest under Wild Growth: %v, want [{G}]", got)
	}
	if got := tapTriggerAdds(t, g, me, fertile); !reflect.DeepEqual(got, []string{"{W|U|B|R|G}"}) {
		t.Errorf("Forest under Fertile Ground: %v, want [{W|U|B|R|G}]", got)
	}
	if got := tapTriggerAdds(t, g, me, plain); !reflect.DeepEqual(got, []string{""}) {
		t.Errorf("unenchanted Forest: %v, want nothing", got)
	}
}

// Mirari's Wake adds "one mana of any type that land produced": on a
// basic it is that land's colour, and on a {W|U} dual it depends on the
// pick, which the field does not carry (the weaker direction).
func TestTapTriggerManaLeavesOutAPickDependentTrigger(t *testing.T) {
	g, me, _ := triggerTable(t)
	pushBattlefieldForTest(g, me.ID, "Mirari's Wake", "Enchantment", plannerWakeOracle)
	forest := pushForest(g, me)
	dual := pushPlannerSource(g, me, "Azorius Land", "Land", "", tapAbility("{W|U}"))
	three := pushPlannerSource(g, me, "Three Picks", "Land", "", tapAbility("{W|U|B|R|G}{W|U|B|R|G}{W|U|B|R|G}"))

	if got := tapTriggerAdds(t, g, me, forest); !reflect.DeepEqual(got, []string{"{G}"}) {
		t.Errorf("Forest under Mirari's Wake: %v, want [{G}]", got)
	}
	if got := tapTriggerAdds(t, g, me, dual); !reflect.DeepEqual(got, []string{""}) {
		t.Errorf("{W|U} dual under Mirari's Wake: %v, want nothing (the extra mana depends on the pick)", got)
	}
	if got := tapTriggerAdds(t, g, me, three); !reflect.DeepEqual(got, []string{""}) {
		t.Errorf("three-pick land: %v, want nothing (above the planner's cap)", got)
	}
}

// CR 106.12: only a tap fires one, as in the planner.
func TestTapTriggerManaNeedsATap(t *testing.T) {
	g, me, _ := triggerTable(t)
	pushBattlefieldForTest(g, me.ID, "Every Tap", "Enchantment", plannerEveryTapOracle)
	gold := pushPlannerSource(g, me, "Gold", "Token Artifact — Gold", "",
		ManaAbilityShape{SacrificeCost: true, Produced: "{W|U|B|R|G}", Label: "Sacrifice: Add one mana of any color"})
	ring := pushPlannerSource(g, me, "Sol Ring", "Artifact", "", tapAbility("{C}{C}"))

	if got := tapTriggerAdds(t, g, me, gold); !reflect.DeepEqual(got, []string{""}) {
		t.Errorf("Gold: %v, want nothing (no tap)", got)
	}
	if got := tapTriggerAdds(t, g, me, ring); !reflect.DeepEqual(got, []string{"{G}"}) {
		t.Errorf("Sol Ring under an every-tap trigger: %v, want [{G}]", got)
	}
}

// CR 106.12b: a tap-keyed doubler doubles the land's mana and not the
// trigger's, which is not a tap (Mana Reflection's ruling).
func TestTapTriggerManaIsPricedAsNotATap(t *testing.T) {
	g, me, _ := triggerTable(t)
	g.testReplacements = append(g.testReplacements, manaProducedReplacement(me.ID, 2, "Test Reflection"))
	pushBattlefieldForTest(g, me.ID, "Crypt Ghast", "Creature — Spirit", plannerGhastOracle)
	swamp := pushLandFor(g, me, "Swamp", "Basic Land — Swamp")

	if got := tapTriggerAdds(t, g, me, swamp); !reflect.DeepEqual(got, []string{"{B}"}) {
		t.Errorf("doubled Swamp under Crypt Ghast: %v, want [{B}]", got)
	}
}

func TestFormatProducedSlotsRoundTrips(t *testing.T) {
	for _, s := range []string{"{B}", "{G}{G}", "{W|U|B|R|G}", "{W3|U3}", "{W|U|B|R|G:2}", "{C}{C}"} {
		slots, err := ParseProducedMana(s)
		if err != nil {
			t.Fatalf("ParseProducedMana(%q): %v", s, err)
		}
		got := formatProducedSlots(slots)
		back, err := ParseProducedMana(got)
		if err != nil {
			t.Fatalf("%q formats as %q, which does not parse: %v", s, got, err)
		}
		if !reflect.DeepEqual(back, slots) {
			t.Errorf("%q formats as %q, which parses to %+v, want %+v", s, got, back, slots)
		}
	}
}

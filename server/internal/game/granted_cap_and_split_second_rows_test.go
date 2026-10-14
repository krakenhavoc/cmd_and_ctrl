package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// #3502 and #3503: the ability rows for a split-second grant and a
// computed mana value cap, and the one reader of the cap.

func TestSplitSecondGrantAndComputedCapRowsAreWorded(t *testing.T) {
	stubRowDefs(t, map[string]*CardDef{
		"rows": {
			SplitSecondGrants: []SplitSecondGrant{{
				Spells: PermissionFilter{InstantOrSorceryOnly: true},
				Label:  "Instant and sorcery spells you control have split second.",
			}},
			GrantedAlternativeCosts: []GrantedAlternativeCost{{
				Offer:            AlternativeCost{},
				Zones:            []ZoneKind{ZoneHand},
				MaxManaValueFunc: func(*Game, uuid.UUID) int { return 0 },
				MaxManaValueText: "the number of creatures you control",
			}},
		},
	})
	got := staticLabels(AbilityRowsOf(Card{OracleID: "rows", Name: "Rows"}))
	want := []string{
		"Instant and sorcery spells you control have split second.",
		"You may cast spells with mana value less than or equal to the number of creatures you control from your hand without paying their mana costs.",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("labels = %q\nwant %q", got, want)
	}
}

// covers reads the fixed cap and the computed one together, and asks the
// computed one for the CASTER it was derived for.
func TestGrantedCoversAppliesBothCapsAndAsksTheCaster(t *testing.T) {
	g := NewGame()
	me, other := uuid.New(), uuid.New()
	counts := map[uuid.UUID]int{me: 4, other: 1}
	two := 2
	gr := GrantedAlternativeCost{
		MaxManaValue:     &two,
		MaxManaValueFunc: func(_ *Game, caster uuid.UUID) int { return counts[caster] },
	}
	three := Card{Name: "Three", TypeLine: "Sorcery", ManaCost: "{2}{U}"}
	two2 := Card{Name: "Two", TypeLine: "Sorcery", ManaCost: "{1}{U}"}
	one := Card{Name: "One", TypeLine: "Sorcery", ManaCost: "{U}"}
	if gr.covers(g, me, three) {
		t.Error("the fixed cap of 2 let a mana value 3 spell through")
	}
	if !gr.covers(g, me, two2) {
		t.Error("a mana value 2 spell under both caps was refused")
	}
	if gr.covers(g, other, two2) {
		t.Error("the computed cap was read for the wrong player")
	}
	if !gr.covers(g, other, one) {
		t.Error("a mana value 1 spell under a computed cap of 1 was refused")
	}
}

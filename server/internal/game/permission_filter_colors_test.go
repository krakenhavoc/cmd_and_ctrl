package game

import "testing"

// ColorsAny is an any-of colour clause read off the card's effective colours
// (#4615): one named colour is enough, a colourless card never matches, and
// it composes with the type clauses as a further AND.
func TestPermissionFilterColorsAnyMatchesAtLeastOneNamedColour(t *testing.T) {
	f := PermissionFilter{CreatureOnly: true, ColorsAny: "RG"}
	for _, tc := range []struct {
		name string
		card Card
		want bool
	}{
		{"red creature", Card{TypeLine: "Creature — Bear", Colors: []string{"R"}}, true},
		{"green creature", Card{TypeLine: "Creature — Bear", Colors: []string{"G"}}, true},
		{"red-blue creature", Card{TypeLine: "Creature — Bear", Colors: []string{"U", "R"}}, true},
		{"blue creature", Card{TypeLine: "Creature — Bear", Colors: []string{"U"}}, false},
		{"colourless creature", Card{TypeLine: "Artifact Creature — Golem"}, false},
		{"red sorcery", Card{TypeLine: "Sorcery", Colors: []string{"R"}}, false},
		{"colour from the cost", Card{TypeLine: "Creature — Bear", ManaCost: "{2}{G}"}, true},
	} {
		if got := f.Matches(tc.card); got != tc.want {
			t.Errorf("%s: Matches = %v, want %v", tc.name, got, tc.want)
		}
	}
	if !(PermissionFilter{}).Matches(Card{TypeLine: "Creature — Bear", Colors: []string{"U"}}) {
		t.Error("the zero filter stopped matching a blue creature")
	}
}

// ColorCountExactly is "that's exactly two colors" (#5645): the number of
// the card's colours, hybrid symbols counted as every colour they show
// (CR 202.2d), and zero constrains nothing.
func TestPermissionFilterColorCountExactlyCountsTheCardsColours(t *testing.T) {
	f := PermissionFilter{InstantOrSorceryOnly: true, ColorCountExactly: 2}
	for _, tc := range []struct {
		name, typeLine, cost string
		want                 bool
	}{
		{"gold instant", "Instant", "{U}{R}", true},
		{"hybrid sorcery", "Sorcery", "{1}{W/B}", true},
		{"hybrid and a third colour", "Instant", "{G}{W/B}", false},
		{"monocolored", "Instant", "{1}{U}{U}", false},
		{"three colours", "Sorcery", "{U}{B}{R}", false},
		{"colorless", "Instant", "{3}", false},
		{"two-colour creature", "Creature — Bear", "{G}{W}", false},
	} {
		c := Card{Name: tc.name, TypeLine: tc.typeLine, ManaCost: tc.cost}
		if got := f.Matches(c); got != tc.want {
			t.Errorf("%s (%s): Matches = %v, want %v", tc.name, tc.cost, got, tc.want)
		}
	}
	if !(PermissionFilter{}).Matches(Card{TypeLine: "Instant", ManaCost: "{3}"}) {
		t.Error("a zero ColorCountExactly constrains nothing")
	}
	if f.faceBlind() {
		t.Error("a colour count is a clause a card's face decides")
	}
}

// DiscardCardInAddition is jump-start's additional cost (CR 702.133a): the
// synthesised offer keeps the card's mana cost and adds one discard.
func TestCastPermissionDiscardCardInAdditionKeepsTheManaCost(t *testing.T) {
	p := &CastPermission{Zone: ZoneGraveyard, AltCostKey: AltCostKeyJumpStart,
		DiscardCardInAddition: true, ExileOnResolution: true}
	ac := p.AlternativeCostFor(Card{Name: "Gold", TypeLine: "Instant", ManaCost: "{U}{R}"})
	if ac == nil || ac.Key != AltCostKeyJumpStart || ac.ManaCost != "{U}{R}" {
		t.Fatalf("offer = %+v, want jump-start at the card's {U}{R}", ac)
	}
	if ac.DiscardFromHand == nil || ac.DiscardFromHand.Min != 1 || ac.DiscardFromHand.Max != 1 {
		t.Errorf("discard component = %+v, want exactly one card from hand", ac.DiscardFromHand)
	}
	if !ac.ExileOnLeavingStack || ac.FromZone != ZoneGraveyard {
		t.Errorf("offer = %+v, want a graveyard cast exiled as it leaves the stack", ac)
	}
}

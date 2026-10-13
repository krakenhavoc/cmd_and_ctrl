package game

import (
	"reflect"
	"testing"
)

// ability_row_labels_test.go — #2410: the structured slots' wording, the
// spent enters-tapped replacement, and the Generic flag.

func stubRowDefs(t *testing.T, defs map[string]*CardDef) {
	t.Helper()
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef { return defs[key] }
	t.Cleanup(func() { CatalogLookup = prev })
}

func staticLabels(rows []AbilityRow) []string {
	var out []string
	for _, r := range rows {
		if r.Kind == AbilityRowStatic {
			out = append(out, r.Label)
		}
	}
	return out
}

func TestStructuredSlotsAreWordedFromTheirFields(t *testing.T) {
	stubRowDefs(t, map[string]*CardDef{
		"slots": {
			HandSize: []HandSizeStatic{
				{Kind: HandSizeNoMaximum},
				{Players: HandSizeEachOpponent, Kind: HandSizeModify, N: -1},
				{Players: HandSizeEachPlayer, Kind: HandSizeNoMaximum},
			},
			AttackLimits: []AttackLimit{
				{Scope: AttackLimitAttackingYou, Max: 2},
				{Scope: AttackLimitEachCombat, Max: 1},
			},
			GameEndGates: []GameEndGate{
				{Scope: GateYou, CantLose: true},
				{Scope: GateOpponents, CantWin: true},
			},
			BlockRules: []BlockRule{
				{Reason: BlockReasonCantBeBlockedBy, Label: "creatures with toughness 3 or greater"},
				{Reason: BlockReasonCantBeBlockedExceptBy, Label: "Walls"},
			},
		},
	})
	got := staticLabels(AbilityRowsOf(Card{OracleID: "slots", Name: "Slots"}))
	want := []string{
		"Can't be blocked by creatures with toughness 3 or greater.",
		"Can't be blocked except by Walls.",
		"No more than 2 creatures can attack you each combat.",
		"No more than 1 creature can attack each combat.",
		"You have no maximum hand size.",
		"Each opponent's maximum hand size is reduced by 1.",
		"Players have no maximum hand size.",
		"You can't lose the game.",
		"Your opponents can't win the game.",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("labels = %q\nwant %q", got, want)
	}
	for _, r := range AbilityRowsOf(Card{OracleID: "slots", Name: "Slots"}) {
		if r.Generic {
			t.Errorf("%q is flagged generic", r.Label)
		}
	}
}

// A replacement that applied as the object entered is spent once the
// object is on the battlefield: the tile in hand still promises it, the
// permanent no longer lists it. A replacement that keeps working is
// listed in both places.
func TestSpentEntryReplacementLeavesTheChipOnceInPlay(t *testing.T) {
	stubRowDefs(t, map[string]*CardDef{
		"tapland": {
			Replacements: []ReplacementEffect{
				{EntryOnly: true, RowLabel: "Enters tapped."},
				{Label: "Doubling: double counters"},
			},
		},
	})
	c := Card{OracleID: "tapland", Name: "Tapland"}
	if got, want := staticLabels(AbilityRowsOf(c)), []string{"Enters tapped.", "Doubling: double counters"}; !reflect.DeepEqual(got, want) {
		t.Errorf("in hand: %q, want %q", got, want)
	}
	if got, want := staticLabels(AbilityRowsOnBattlefield(c)), []string{"Doubling: double counters"}; !reflect.DeepEqual(got, want) {
		t.Errorf("in play: %q, want %q", got, want)
	}
	// A land whose only ability was the enters-tapped replacement has no
	// ◆ chip at all once it is in play.
	stubRowDefs(t, map[string]*CardDef{"bare": {Replacements: []ReplacementEffect{{EntryOnly: true, RowLabel: "Enters tapped."}}}})
	if rows := AbilityRowsOnBattlefield(Card{OracleID: "bare", Name: "Bare"}); len(rows) != 0 {
		t.Errorf("a tapland in play lists %+v, want none", rows)
	}
}

// RowLabel is the chip's wording; Label stays the prompt header and is
// only the fallback.
func TestReplacementRowLabelWinsOverTheHeader(t *testing.T) {
	stubRowDefs(t, map[string]*CardDef{
		"r": {Replacements: []ReplacementEffect{{Label: "Shock Land", RowLabel: "Enters tapped unless you pay 2 life."}, {Label: "Shock Land"}}},
	})
	rows := AbilityRowsOf(Card{OracleID: "r", Name: "Shock Land"})
	if len(rows) != 2 || rows[0].Label != "Enters tapped unless you pay 2 life." || rows[0].Generic {
		t.Fatalf("rows = %+v", rows)
	}
	// A Label that is only the card's own name says nothing: generic.
	if !rows[1].Generic || rows[1].Label != "Replacement effect" {
		t.Errorf("name-only label row = %+v, want the generic fallback", rows[1])
	}
}

func TestGenericFlagMarksKindOnlyRows(t *testing.T) {
	stubRowDefs(t, map[string]*CardDef{
		"g": {Static: []StaticAbility{{Layer: Layer7PT}, {Layer: Layer7PT, Label: "Gets +1/+1."}}},
	})
	rows := AbilityRowsOf(Card{OracleID: "g", Name: "G"})
	if len(rows) != 2 || !rows[0].Generic || rows[1].Generic {
		t.Errorf("rows = %+v, want generic then worded", rows)
	}
}

func TestCastPermissionLabelDropsTheTrailingCardName(t *testing.T) {
	stubRowDefs(t, map[string]*CardDef{
		"p": {CastPermissions: []CastPermission{{Label: "Play a land from your graveyard (Crucible of Worlds)"}}},
	})
	got := staticLabels(AbilityRowsOf(Card{OracleID: "p", Name: "Crucible of Worlds"}))
	if want := []string{"Play a land from your graveyard"}; !reflect.DeepEqual(got, want) {
		t.Errorf("labels = %q, want %q", got, want)
	}
}

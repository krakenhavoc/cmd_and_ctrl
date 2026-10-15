package game

import (
	"testing"
)

// trigger_counted_test.go — #3033: an X the engine stamps from its own
// cast record (TriggeredAbility.XFromCast) is plain: the row describes
// the item. The same XValue on a row without the flag is something the
// check cannot read.
func TestAnXStampedFromTheCastIsPlainOnlyForAnXFromCastRow(t *testing.T) {
	for _, fromCast := range []bool{true, false} {
		row := TriggeredAbility{
			Key:       "probe",
			Effect:    func(*Game, *StackItem) error { return nil },
			Footprint: []FootprintStep{{Kind: FootprintCreateToken, N: 1}},
			XFromCast: fromCast,
		}
		old := CatalogTriggers
		CatalogTriggers = func(key string) []TriggeredAbility {
			if key == "x-probe" {
				return []TriggeredAbility{row}
			}
			return nil
		}
		f := newCondFixture(t)
		item := f.item()
		item.Body = CatalogTriggeredBodyKey
		item.Params.Ability = &AbilityRef{Key: "x-probe", Slot: AbilitySlotTriggered, Name: "probe", Ref: OwnAbilityRef(0)}
		item.XValue = 3
		_, ok := f.g.triggerAccessLocked(item)
		CatalogTriggers = old
		if ok != fromCast {
			t.Errorf("XFromCast %v: described %v", fromCast, ok)
		}
	}
}

// An XFromCast row's built item carries its source's cast X.
func TestAnXFromCastRowStampsTheSourcesX(t *testing.T) {
	g := newActiveGame(t)
	src := Card{Name: "X Probe", TypeLine: "Enchantment"}
	src.Provenance.X = 4
	row := TriggeredAbility{Key: "probe", Effect: func(*Game, *StackItem) error { return nil }, XFromCast: true}
	if item := g.buildTriggerItemLocked(row, Event{}, src, Characteristic{}); item == nil || item.XValue != 4 {
		t.Errorf("item %+v, want XValue 4", item)
	}
	row.XFromCast = false
	if item := g.buildTriggerItemLocked(row, Event{}, src, Characteristic{}); item == nil || item.XValue != 0 {
		t.Errorf("without the flag: item %+v, want XValue 0", item)
	}
}

package game

import (
	"testing"

	"github.com/google/uuid"
)

// backup_test.go — #5651: the keyword a targeting triggered ability is
// (Event.TargeterKeyword), read through the catalog row its stack item
// names.

func withCatalogTriggerRows(t *testing.T, key string, rows []TriggeredAbility) {
	t.Helper()
	prev := CatalogTriggers
	def := &CardDef{Triggered: rows}
	IdentifyCatalogRows(key, def)
	CatalogTriggers = func(k string) []TriggeredAbility {
		if k == key {
			return def.Triggered
		}
		if prev != nil {
			return prev(k)
		}
		return nil
	}
	t.Cleanup(func() { CatalogTriggers = prev })
}

// TestTargeterKeywordReadsTheCatalogRow: only a triggered item that
// names a catalog row whose constructor stamped a keyword answers.
func TestTargeterKeywordReadsTheCatalogRow(t *testing.T) {
	const key = "test-backup-creature"
	noop := func(*Game, *StackItem) error { return nil }
	withCatalogTriggerRows(t, key, []TriggeredAbility{
		{Key: "backup row", Keyword: KeywordBackup, Effect: noop},
		{Key: "plain row", Effect: noop},
	})
	rows := CatalogTriggers(key)
	item := func(row int) *StackItem {
		ref := TriggeredAbilityRef(rows[row])
		if ref == nil {
			t.Fatalf("row %d has no catalog ref", row)
		}
		return &StackItem{Kind: StackItemTriggered, Body: CatalogTriggeredBodyKey, Params: EffectParams{Ability: ref}}
	}
	lost := item(0)
	lost.Params.Ability = &AbilityRef{Key: key, Slot: AbilitySlotTriggered, Ref: "own:7", Name: "a row that is gone"}
	activated := item(0)
	activated.Kind = StackItemActivated

	for name, c := range map[string]struct {
		item *StackItem
		want string
	}{
		"backup row":          {item(0), KeywordBackup},
		"plain row":           {item(1), ""},
		"no item":             {nil, ""},
		"spell":               {&StackItem{Kind: StackItemSpell}, ""},
		"unkeyed trigger":     {&StackItem{Kind: StackItemTriggered}, ""},
		"not a trigger":       {activated, ""},
		"row no longer there": {lost, ""},
	} {
		if got := targeterKeyword(c.item); got != c.want {
			t.Errorf("%s: targeterKeyword = %q, want %q", name, got, c.want)
		}
	}
}

// TestBecomesTargetCarriesTheTargetersKeyword: the event is stamped from
// the item as it is emitted, a copy of the item reads the same, and the
// field rides a clone. The disk restore is played with the real cards
// (TestMirrorShieldHopliteSeesABackupAbilityAcrossARestore).
func TestBecomesTargetCarriesTheTargetersKeyword(t *testing.T) {
	const key = "test-backup-creature"
	withCatalogTriggerRows(t, key, []TriggeredAbility{
		{Key: "backup row", Keyword: KeywordBackup, Effect: func(*Game, *StackItem) error { return nil }},
	})
	g := newActiveGame(t)
	me := g.Seats[0]
	victim := crimeTestCard(g, g.Battlefield, "Victim", me.ID, me.ID)
	ref := TriggeredAbilityRef(CatalogTriggers(key)[0])
	if ref == nil {
		t.Fatal("the row has no catalog ref")
	}
	g.WithWriteLock(func() {
		id := uuid.New()
		item := &StackItem{ID: id, Kind: StackItemTriggered, Controller: me.ID, SourceCardID: uuid.New(),
			Body: CatalogTriggeredBodyKey, Params: EffectParams{Ability: ref}}
		g.StackMeta = map[uuid.UUID]*StackItem{id: item}
		target := []TargetRef{{Kind: TargetCard, ID: victim}}
		g.emitBecameTargetLocked(me.ID, item.SourceCardID, id, target)
		// CR 707.10: the copy names the same row, so it is a backup
		// ability too, and its own became-target event says so.
		g.createAbilityCopyLocked(item, me.ID, target)
	})
	check := func(name string, evs []Event) {
		t.Helper()
		if len(evs) != 2 {
			t.Fatalf("%s: %d became-target events, want 2 (the ability and its copy)", name, len(evs))
		}
		for i, ev := range evs {
			if ev.TargeterKind != StackItemTriggered || ev.TargeterKeyword != KeywordBackup {
				t.Errorf("%s event %d: kind %q keyword %q, want a triggered backup ability", name, i, ev.TargeterKind, ev.TargeterKeyword)
			}
		}
	}
	check("live", becameTargetEvents(g))
	check("clone", becameTargetEvents(g.Clone()))
}

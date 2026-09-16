package game

import (
	"testing"

	"github.com/google/uuid"
)

// carddef_test.go — #622: the per-slot hooks are views onto
// CatalogLookup, and a nil lookup means no catalog.

func TestPerSlotHooksDeriveFromCatalogLookup(t *testing.T) {
	prev := CatalogLookup
	t.Cleanup(func() { CatalogLookup = prev })

	CatalogLookup = nil
	if IsCatalogCard("x") || CatalogTriggers("x") != nil || CatalogStaticAbilities("x") != nil || CatalogStartingLoyalty("x") != 0 || CatalogCantBeCountered("x") {
		t.Fatal("with no lookup wired every slot must read as absent")
	}
	if err := EffectResolver(nil, &StackItem{}, "x"); err != nil {
		t.Fatalf("EffectResolver with no catalog: %v", err)
	}

	def := &CardDef{
		StartingLoyalty: 3,
		CantBeCountered: true,
		Triggered:       []TriggeredAbility{{Watches: []EventKind{EventETB}}},
		PrintedKeywords: []string{"flying"},
		CastableZones:   []ZoneKind{ZoneGraveyard},
	}
	CatalogLookup = func(key string) *CardDef {
		if key == "probe" {
			return def
		}
		return nil
	}
	if !IsCatalogCard("probe") || IsCatalogCard("other") || IsCatalogCard("") {
		t.Error("IsCatalogCard must follow the lookup, and an empty key is never a card")
	}
	if got := CatalogTriggers("probe"); len(got) != 1 || &got[0] != &def.Triggered[0] {
		t.Error("CatalogTriggers must hand out the def's own slice")
	}
	if CatalogStartingLoyalty("probe") != 3 || !CatalogCantBeCountered("probe") || len(CatalogCastableZones("probe")) != 1 || CatalogPrintedKeywords("probe")[0] != "flying" {
		t.Error("scalar and slice slots must read through")
	}
	resolved := false
	def.Resolve = func(*Game, *StackItem) error { resolved = true; return nil }
	if err := EffectResolver(nil, &StackItem{}, "probe"); err != nil || !resolved {
		t.Error("EffectResolver must run the def's Resolve")
	}
	entered := uuid.Nil
	def.AsEnters = func(_ *Game, id uuid.UUID) error { entered = id; return nil }
	want := uuid.New()
	if err := ETBEffectHook(nil, want, "probe"); err != nil || entered != want {
		t.Error("ETBEffectHook must run the def's AsEnters with the card ID")
	}
}

package game

import "testing"

// TestSacrificeKindNamesNarrowTheKind pins SacrificeKind.Names (S82 OLD batch
// 800, Urborg Panther's "a creature named Feral Shadow"): a named kind is of
// that name AND its type words, and a kind with no names reads as it always did.
func TestSacrificeKindNamesNarrowTheKind(t *testing.T) {
	shadow := Card{Name: "Feral Shadow", TypeLine: "Creature — Nightstalker"}
	other := Card{Name: "Bear", TypeLine: "Creature — Bear"}
	namedLand := Card{Name: "Feral Shadow", TypeLine: "Land"}
	k := SacrificeKind{Label: "a creature named Feral Shadow", CardTypes: []string{"creature"}, Names: []string{"Feral Shadow"}}
	if !k.Matches(shadow) {
		t.Error("the named creature does not match")
	}
	if k.Matches(other) {
		t.Error("a creature of another name matches")
	}
	if k.Matches(namedLand) {
		t.Error("a permanent of the name but another type matches")
	}
	nameOnly := SacrificeKind{Label: "a permanent named Feral Shadow", Names: []string{"Feral Shadow"}}
	if !nameOnly.Matches(namedLand) || nameOnly.Matches(other) {
		t.Error("a name-only kind should match by name alone")
	}
	plain := SacrificeKind{Label: "a creature", CardTypes: []string{"creature"}}
	if !plain.Matches(other) {
		t.Error("a kind with no names stopped matching by type")
	}
}

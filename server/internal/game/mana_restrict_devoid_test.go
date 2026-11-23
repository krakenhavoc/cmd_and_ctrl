package game

import "testing"

// mana_restrict_devoid_test.go — #5948: "Spend this mana only to cast a spell
// with devoid" (Corrupted Crossroads), the ManaRestrictDevoid tag.

func TestManaRestrictDevoidPaysOnlyForCastsOfSpellsWithDevoid(t *testing.T) {
	devoid := Card{Name: "Devoid Drone", TypeLine: "Creature — Eldrazi Drone", ManaCost: "{1}{U}", Keywords: []string{KeywordDevoid}}
	plain := Card{Name: "Plain Drone", TypeLine: "Creature — Eldrazi Drone", ManaCost: "{1}{U}", Colors: []string{"U"}}
	colorlessArtifact := Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{2}"}
	tag := []string{ManaRestrictDevoid}
	if !ManaSpendForCast(devoid).allows(tag) {
		t.Error("mana restricted to devoid spells can't cast a spell with devoid")
	}
	if ManaSpendForCast(plain).allows(tag) {
		t.Error("mana restricted to devoid spells cast a coloured spell")
	}
	if ManaSpendForCast(colorlessArtifact).allows(tag) {
		t.Error("mana restricted to devoid spells cast a colorless artifact, which does not have devoid")
	}
	if ManaSpendForAbility(devoid).allows(tag) {
		t.Error("mana restricted to casting a spell paid for an ability")
	}
	if (ManaSpendContext{}).allows(tag) {
		t.Error("an unknown payment satisfied the restriction")
	}
}

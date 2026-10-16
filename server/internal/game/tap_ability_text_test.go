package game

import "testing"

// TestPrintsTapAbilityText pins the text scan behind
// Card.PrintsTapAbility (#3860, Magewright's Stone).
func TestPrintsTapAbilityText(t *testing.T) {
	cases := []struct {
		name, typeLine, text string
		want                 bool
	}{
		{"mana dork", "Creature — Elf Druid", "{T}: Add {G}.", true},
		{"with a mana cost too", "Creature — Human", "{1}, {T}: Draw a card, then discard a card.", true},
		{"second line", "Creature — Human", "Flying\n{T}, Sacrifice a creature: Deal 1 damage to any target.", true},
		{"ability word prefix", "Creature — Human", "Boast — {T}: Draw a card.", true},
		{"no tap in the cost", "Creature — Human", "{2}: Gain 1 life.", false},
		{"tap in the effect only", "Creature — Human", "{2}: Tap target creature. {T}", false},
		{"reminder text only", "Basic Land — Forest", "({T}: Add {G}.)", false},
		{"granted in quotes", "Creature — Human", "Creatures you control have \"{T}: Add {G}.\"", false},
		{"equipped creature has", "Artifact — Equipment", "Equipped creature has \"{T}: Add {C}.\"\nEquip {2}", false},
		{"instant", "Instant", "{T}: Add {G}.", false},
		{"vanilla keyword", "Creature — Bear", "Trample", false},
		{"loyalty cost", "Planeswalker — Jace", "+1: Draw a card.", false},
		{"empty", "Creature — Bear", "", false},
	}
	for _, tc := range cases {
		if got := PrintsTapAbilityText(tc.typeLine, tc.text); got != tc.want {
			t.Errorf("%s: PrintsTapAbilityText = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// A face-down permanent has no text (CR 708.2a), so no printed tap
// ability, whatever is underneath.
func TestPrintedTapAbilityIsFalseFaceDown(t *testing.T) {
	c := Card{TypeLine: "Creature — Elf", PrintsTapAbility: true}
	if !c.PrintedTapAbility() {
		t.Fatal("setup: a face-up card with the stamp reads true")
	}
	c.SetFaceDown(FaceDownMorphed)
	if c.PrintedTapAbility() {
		t.Error("a face-down permanent has no text")
	}
}

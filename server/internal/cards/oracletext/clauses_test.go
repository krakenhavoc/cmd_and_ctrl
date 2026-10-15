package oracletext

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestClausesSortsTheParagraphs(t *testing.T) {
	text := strings.Join([]string{
		"Flying",
		"Landfall — Whenever a land you control enters, draw a card.",
		"{T}: Choose one —",
		"• Add {G}.",
		"• Scry 1.",
		"You may exert this creature as it attacks. When you do, it gets +1/+1 until end of turn.",
		"As this enchantment enters, choose Khans or Dragons.",
		"• Khans — At the beginning of combat on your turn, draw a card.",
		"• Dragons — Spells you cast cost {1} less.",
		"I, II — Create a 1/1 white Soldier creature token.",
	}, "\n")
	got := Clauses(text, CardNames("Test Card"))
	want := []struct {
		kind ClauseKind
		text string
	}{
		{ClauseOther, "Flying"},
		{ClauseTriggered, "Landfall — Whenever a land you control enters, draw a card."},
		{ClauseActivated, "{T}: Choose one —\n• Add {G}.\n• Scry 1."},
		{ClauseTriggered, "You may exert this creature as it attacks. When you do, it gets +1/+1 until end of turn."},
		{ClauseOther, "As this enchantment enters, choose Khans or Dragons.\n• Dragons — Spells you cast cost {1} less."},
		{ClauseTriggered, "Khans — At the beginning of combat on your turn, draw a card."},
		{ClauseTriggered, "I, II — Create a 1/1 white Soldier creature token."},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d clauses, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].Kind != w.kind || got[i].Text != w.text {
			t.Errorf("clause %d = (%d, %q), want (%d, %q)", i, got[i].Kind, got[i].Text, w.kind, w.text)
		}
	}
}

func TestMatchTriggered(t *testing.T) {
	names := CardNames("Archangel of Wrath")
	clauses := Clauses(strings.Join([]string{
		"Kicker {B} and/or {R} (You may pay an additional {B} and/or {R} as you cast this spell.)",
		"Echo {3}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)",
		"When this creature enters, if it was kicked, it deals 2 damage to any target.",
		"When this creature enters, if it was kicked twice, it deals 2 damage to any target.",
		"Draw a card.",
	}, "\n"), names)
	got := MatchTriggered([]string{
		"Archangel of Wrath — kicked, 2 damage to any target",
		"Archangel of Wrath — kicked twice, 2 damage to any target",
		"Archangel of Wrath — echo {3}{R}",
		// A keyword line with no reminder text is not a trigger's text.
		"Archangel of Wrath — draw a card",
		// No clause says it.
		"Archangel of Wrath — gain 3 life",
	}, clauses, names)
	want := []string{
		// Tied with the "kicked twice" clause, which the next row claims.
		"When this creature enters, if it was kicked, it deals 2 damage to any target.",
		"When this creature enters, if it was kicked twice, it deals 2 damage to any target.",
		"Echo {3}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)",
		"",
		"",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d = %q, want %q", i, got[i], want[i])
		}
	}
}

// Two rows that say the same thing, between two clauses that both say
// it, are not guessed.
func TestMatchTriggeredLeavesAnEvenTie(t *testing.T) {
	names := CardNames("Seraph Sanctuary")
	clauses := Clauses("When this land enters, you gain 1 life.\nWhenever an Angel you control enters, you gain 1 life.", names)
	got := MatchTriggered([]string{"Seraph Sanctuary — you gain 1 life", "Seraph Sanctuary — you gain 1 life"}, clauses, names)
	if got[0] != "" || got[1] != "" {
		t.Errorf("got %q, want no answer for either", got)
	}
}

func TestMatchActivated(t *testing.T) {
	names := CardNames("Purphoros, God of the Forge")
	clauses := Clauses("Equip {2}\n{2}{R}: Creatures you control get +1/+0 until end of turn.\n−3: Exile target permanent.", names)
	got := MatchActivated([]string{
		"{2}{R}: Creatures you control get +1/+0 until end of turn.",
		"Equip {2}",
		"-3: Exile target permanent.",
		"{1}: Draw a card.",
	}, clauses, names)
	want := []string{
		"{2}{R}: Creatures you control get +1/+0 until end of turn.",
		"Equip {2}",
		"−3: Exile target permanent.",
		"",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("row %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestAbilityTextWithNoSourceIsEmpty(t *testing.T) {
	SetSource(nil)
	ref := game.AbilityRef{Key: "4fdbbec2-e921-4b63-958d-f9ba1e417197", Slot: game.AbilitySlotTriggered, Ref: game.OwnAbilityRef(0), Name: "x"}
	if got := AbilityText(ref); got != "" {
		t.Errorf("got %q with no source, want none", got)
	}
	if !strings.HasPrefix(game.OwnAbilityRef(3), ownRefPrefix) {
		t.Errorf("game.OwnAbilityRef(3) = %q, not %q-prefixed", game.OwnAbilityRef(3), ownRefPrefix)
	}
}

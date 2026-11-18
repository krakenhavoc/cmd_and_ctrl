package deck

import (
	"fmt"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gyruda is Gyruda, Doom of Depths as a companion card.
func gyruda() cards.Card {
	c := basicLegal("Gyruda, Doom of Depths", "Legendary Creature — Demon Kraken", "U", "B")
	c.OracleID = uuid.MustParse(gyrudaOracleID)
	c.ManaCost = "{4}{U/B}{U/B}"
	c.Keywords = []string{"Companion"}
	c.OracleText = "Companion — Your starting deck contains only cards with even mana values. (If this card is your chosen companion, you may put it into your hand from outside the game for {3} as a sorcery.)"
	return c
}

// gyrudaDeck is a legal two-colour deck whose every card has an even mana
// value, with Gyruda as its companion; extra replaces that many of the 99.
func gyrudaDeck(extra ...cards.Card) *List {
	cmd := costed("Blue Black Commander", "Legendary Creature — Wizard", "{2}{U}{B}", "U", "B")
	list := &List{Commanders: []cards.Card{cmd}}
	for i := range 99 - len(extra) {
		list.Mainboard = append(list.Mainboard, costed(fmt.Sprintf("Spell %d", i), "Instant", "{1}{U}", "U"))
	}
	list.Mainboard = append(list.Mainboard, extra...)
	g := gyruda()
	list.Companion = &g
	return list
}

func TestGyrudaCompanionConditionEdgeCases(t *testing.T) {
	split := costed("Fire // Ice", "Instant // Instant", "{1}{R} // {1}{U}", "U", "R")
	split.Layout = "split"
	split.CardFaces = []cards.CardFace{{Name: "Fire", ManaCost: "{1}{R}"}, {Name: "Ice", ManaCost: "{1}{U}"}}

	oddSplit := costed("Odd // Even", "Instant // Instant", "{1}{U} // {U}", "U")
	oddSplit.Layout = "split"
	oddSplit.CardFaces = []cards.CardFace{{Name: "Odd", ManaCost: "{1}{U}"}, {Name: "Even", ManaCost: "{U}"}}

	dfc := costed("Valki, God of Lies // Tibalt, Cosmic Impostor", "Legendary Creature — God // Legendary Planeswalker — Tibalt", "", "B", "R")
	dfc.Layout = "modal_dfc"
	dfc.CardFaces = []cards.CardFace{{Name: "Valki, God of Lies", ManaCost: "{1}{B}"}, {Name: "Tibalt, Cosmic Impostor", ManaCost: "{5}{B}{R}"}}

	oddDFC := costed("Odd Front // Even Back", "Creature // Creature", "", "B")
	oddDFC.Layout = "transform"
	oddDFC.CardFaces = []cards.CardFace{{Name: "Odd Front", ManaCost: "{2}{B}"}, {Name: "Even Back", ManaCost: ""}}

	cases := []struct {
		name  string
		card  cards.Card
		fails bool
	}{
		{"odd mana value", costed("Dark Ritual-ish", "Instant", "{B}", "B"), true},
		{"even mana value", costed("Counterspell", "Instant", "{U}{U}", "U"), false},
		{"X counts as 0", costed("Even X", "Sorcery", "{X}{U}{U}", "U"), false},
		{"X with an odd cost", costed("Odd X", "Sorcery", "{X}{U}", "U"), true},
		{"hybrid takes its largest component", costed("Twobrid", "Creature", "{2/B}{U}", "B", "U"), true},
		{"hybrid pair is even", costed("Kitchen Finks", "Creature", "{1}{G/W}{G/W}", "G", "W"), true},
		{"phyrexian symbol is one", costed("Gitaxian-ish", "Instant", "{U/P}{U}", "U"), false},
		{"split card halves combine, even", split, false},
		{"split card halves combine, odd", oddSplit, true},
		{"double-faced card is its front face, even", dfc, false},
		{"double-faced card is its front face, odd", oddDFC, true},
		{"land is zero", basicLegal("Command Tower", "Land"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(gyrudaDeck(tc.card), game.FormatCommander)
			failed := slices.Contains(violationCodes(err), CodeCompanionConditionNotMet)
			if failed != tc.fails {
				t.Errorf("condition failed = %v, want %v (%v)", failed, tc.fails, err)
			}
		})
	}
}

func TestGyrudaCompanionDeckThatFulfilsItValidates(t *testing.T) {
	if err := Validate(gyrudaDeck(), game.FormatCommander); err != nil {
		t.Fatalf("a deck of even mana values with Gyruda: %v", err)
	}
}

// CR 702.139b: the starting deck includes the commander.
func TestGyrudaCompanionConditionCountsTheCommander(t *testing.T) {
	list := gyrudaDeck()
	list.Commanders[0].ManaCost = "{1}{U}{B}"
	if codes := violationCodes(Validate(list, game.FormatCommander)); !slices.Contains(codes, CodeCompanionConditionNotMet) {
		t.Errorf("an odd commander must break the condition: %v", codes)
	}
}

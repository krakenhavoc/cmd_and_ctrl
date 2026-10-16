package deck

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// TestImporterStampsPrintsTapAbility: the importer reads the printed
// text once (#3860), so Magewright's Stone can target an uncatalogued
// creature that has a {T} ability. A multi-faced card is judged on the
// face that is up, which SetFace(0) settles at import.
func TestImporterStampsPrintsTapAbility(t *testing.T) {
	elf := cards.Card{Name: "Fake Elf", Layout: "normal", TypeLine: "Creature — Elf Druid", OracleText: "{T}: Add {G}."}
	dfc := cards.Card{
		Name: "Fake Human // Fake Wolf", Layout: "transform", TypeLine: "Creature — Human // Creature — Werewolf",
		CardFaces: []cards.CardFace{
			{Name: "Fake Human", TypeLine: "Creature — Human", OracleText: "Vigilance", Power: "1", Toughness: "1"},
			{Name: "Fake Wolf", TypeLine: "Creature — Werewolf", OracleText: "{T}: Add {G}.", Power: "2", Toughness: "2"},
		},
	}
	cases := []struct {
		name string
		card cards.Card
		want bool
	}{
		{"mana elf", elf, true},
		{"vanilla creature", grizzlyBears(), false},
		{"transform card, front face has none", dfc, false},
	}
	for _, tc := range cases {
		if got := toGameCard(tc.card, false).PrintsTapAbility; got != tc.want {
			t.Errorf("%s: PrintsTapAbility = %v, want %v", tc.name, got, tc.want)
		}
	}
	back := toGameCard(dfc, false)
	back.SetFace(1)
	if !back.PrintedTapAbility() {
		t.Error("the back face prints a tap ability and is up")
	}
}

package deck

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestImporterStampsManaAbilityText: the importer reads the printed text
// once for a mana ability as CR 605.1 defines one (#3943), so Moonsilver
// Key can find an artifact the catalog does not run, and does not find
// one whose only mana is a dies trigger. A double-faced card is judged on
// its front face, which is the card in a library (CR 712.8a).
func TestImporterStampsManaAbilityText(t *testing.T) {
	rock := cards.Card{Name: "Fake Rock", Layout: "normal", TypeLine: "Artifact", OracleText: "{T}: Add {C}{C}.", ProducedMana: []string{"C"}}
	suchi := cards.Card{Name: "Fake Su-Chi", Layout: "normal", TypeLine: "Artifact Creature — Construct", OracleText: "When this creature dies, add {C}{C}{C}{C}.", ProducedMana: []string{"C"}}
	dfc := cards.Card{
		Name: "Fake Amulet // Fake Wellspring", Layout: "transform", TypeLine: "Artifact // Land",
		CardFaces: []cards.CardFace{
			{Name: "Fake Amulet", TypeLine: "Artifact", OracleText: "Instant and sorcery spells you cast cost {1} less to cast."},
			{Name: "Fake Wellspring", TypeLine: "Land", OracleText: "{T}: Add one mana of any color."},
		},
	}
	cases := []struct {
		name string
		card cards.Card
		want game.PrintedFact
	}{
		{"a mana rock", rock, game.PrintedYes},
		{"a dies trigger is not a mana ability", suchi, game.PrintedNo},
		{"a vanilla creature", grizzlyBears(), game.PrintedNo},
		{"a transform card judged on its front face", dfc, game.PrintedNo},
	}
	for _, tc := range cases {
		if got := toGameCard(tc.card, false).ManaAbilityText; got != tc.want {
			t.Errorf("%s: ManaAbilityText = %v, want %v", tc.name, got, tc.want)
		}
	}
	back := toGameCard(dfc, false)
	back.SetFace(1)
	if back.ManaAbilityText != game.PrintedYes {
		t.Error("the back face has a mana ability and is up")
	}
}

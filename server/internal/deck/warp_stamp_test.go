package deck

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestImporterStampsWarpText: the importer reads the printed text once
// for the warp keyword (#5688), so Blade of the Swarm can find an exiled
// card with warp that the catalog does not run. A card that only grants
// warp does not have it, and a double-faced card is judged on the face
// that is up.
func TestImporterStampsWarpText(t *testing.T) {
	granter := cards.Card{
		Name: "Fake Tannuk", Layout: "normal", TypeLine: "Legendary Creature — Kavu Pilot", Keywords: []string{"Warp"},
		OracleText: "Artifact cards and red creature cards in your hand have warp {2}{R}. (You may cast a card from your hand for its warp cost.)",
	}
	dfc := cards.Card{
		Name: "Fake Warper // Fake Husk", Layout: "transform", TypeLine: "Creature — Drix // Creature — Husk", Keywords: []string{"Warp"},
		CardFaces: []cards.CardFace{
			{Name: "Fake Warper", TypeLine: "Creature — Drix", OracleText: "Warp {1}{U}"},
			{Name: "Fake Husk", TypeLine: "Creature — Husk", OracleText: "Flying"},
		},
	}
	cases := []struct {
		name string
		card cards.Card
		want game.PrintedFact
	}{
		{"a card with warp", anticausalVestige(), game.PrintedYes},
		{"a card that grants warp", granter, game.PrintedNo},
		{"a vanilla creature", grizzlyBears(), game.PrintedNo},
		{"a transform card with warp on its front face", dfc, game.PrintedYes},
	}
	for _, tc := range cases {
		if got := toGameCard(tc.card, false).WarpText; got != tc.want {
			t.Errorf("%s: WarpText = %v, want %v", tc.name, got, tc.want)
		}
	}
	back := toGameCard(dfc, false)
	back.SetFace(1)
	if back.WarpText != game.PrintedNo {
		t.Error("the back face prints no warp and is up")
	}
}

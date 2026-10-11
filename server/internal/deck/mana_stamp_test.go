package deck

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// TestImporterStampsPrintsManaAbility: the importer stamps the mana
// text scan (#2258); a basic land's reminder text does not count.
func TestImporterStampsPrintsManaAbility(t *testing.T) {
	rock := cards.Card{Name: "Fake Rock", Layout: "normal", TypeLine: "Artifact", OracleText: "{T}: Add {C}{C}."}
	cases := []struct {
		name string
		card cards.Card
		want bool
	}{
		{"mana rock", rock, true},
		{"basic land", island(), false},
		{"vanilla creature", grizzlyBears(), false},
	}
	for _, tc := range cases {
		if got := toGameCard(tc.card, false).PrintsManaAbility; got != tc.want {
			t.Errorf("%s: PrintsManaAbility = %v, want %v", tc.name, got, tc.want)
		}
	}
}

package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tokens_297a_more_test.go — the rest of slice 297-a: Skyclave
// Apparition, Overlord of the Hauntwoods, Ondu Spiritdancer, Myr
// Battlesphere and Rain of Riches.

const (
	skyclaveApparitionOracle = "d90af00a-d322-4265-9954-7b1e80702e18"
	OverlordHauntwoodsOracle = "4669e8c7-fc37-4b97-9cb7-8d29b43d9176"
	OnduSpiritdancerOracle   = "d2ae893f-52af-4645-8712-d62fb78f901e"
	MyrBattlesphereOracle    = "c53ba31a-ba27-4e17-9a92-311acb1cab29"
	RainOfRichesOracle       = "ef1e2d3d-e977-4729-9430-eba6242e5dfe"
)

// --- Skyclave Apparition ------------------------------------------

func CastSkyclaveApparition(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := CastFromHand(t, g, game.Card{
		Name: "Skyclave Apparition", TypeLine: "Creature — Kor Spirit", OracleID: skyclaveApparitionOracle,
		ManaCost: "{1}{W}{W}", Power: 2, Toughness: 2,
	}, game.CastSpellParams{})
	PassPriorityAroundTable(t, g)
	return id
}

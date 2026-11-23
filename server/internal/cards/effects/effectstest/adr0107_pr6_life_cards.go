package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0107_pr6_life_cards_test.go — ADR 0107 §5 (#1880): the "can't
// gain life" cards, and Sulfuric Vortex's and Screaming Nemesis's
// clauses that waited on the same seam.

const (
	ArchfiendOfDespairOracle = "9c36760b-57c5-488b-afc9-ef141942c6ab"
	AtarkasCommandOracle     = "1943cbb0-2b12-4450-90af-060b8c8627c3"
	ErebosGodOfTheDeadOracle = "1cfc7b12-a595-492d-81ee-bd100ba7de6a"
	ForsakenWastesOracle     = "b9e61e68-9dc8-4295-95dc-dd66a0907c8c"
	GiantCindermawOracle     = "bd0c0e85-ff24-46c9-a547-4f633355c131"
	GrievousWoundOracle      = "4d97b9d1-1138-40da-b593-8065ea6ed7b3"
	GrimaWormtongueOracle    = "8f98806e-c6b7-44af-b436-88acf36d25ec"
	HavocFestivalOracle      = "af0dd100-67f0-4d7c-a65b-b2c7149288f7"
	KnightOfDusksOracle      = "815ae30f-f455-4795-96eb-4bcc41d415e8"
	QuakebringerOracle       = "bbe41a21-f8ad-4e5c-88d8-2a41dfd32d13"
	RampagingFerocidonOracle = "3e5ca524-8dd1-4f7f-a467-5210d768b8a1"
	RoilingVortexOracle      = "34f83b7a-45a7-40a8-abcb-31d7aae4cea1"
	StigmaLasherOracle       = "7e4a1286-cac6-491f-a32d-64fb98d35de2"
	TheLordOfPainOracle      = "b194f277-5045-4391-8b48-d555a6d657a7"
	TibaltRakishOracle       = "0b768f8f-2213-45b9-bced-8fb1bbb441c3"
	SulfuricVortexOracle     = "7652f328-e142-494b-a869-772ced10c26a"
	ScreamingNemesisOracle   = "fcb7c93c-46ab-49b5-a6e0-35d73f3be8f0"
)

// gains reports whether `player` gains life from a 3-point sandbox gain,
// leaving the total where it was found.
func Gains(t *testing.T, g *game.Game, player uuid.UUID) bool {
	t.Helper()
	before := LifeTotalOf(g, player)
	if _, err := g.ChangePlayerLife(player, 3); err != nil {
		t.Fatalf("ChangePlayerLife: %v", err)
	}
	after := LifeTotalOf(g, player)
	if after != before {
		if _, err := g.ChangePlayerLife(player, before-after); err != nil {
			t.Fatalf("ChangePlayerLife: %v", err)
		}
	}
	return after > before
}

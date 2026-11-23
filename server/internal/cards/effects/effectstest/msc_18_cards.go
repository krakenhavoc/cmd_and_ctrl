package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_18_cards_test.go — Marvel Super Heroes Commander, A–L, batch 18 (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	Msc18ArchnemesisOracle   = "c408717c-e07b-484f-9c81-5666595adf81"
	Msc18CrucibleOracle      = "2469c0b7-a045-4733-bf53-28f3fdbe3797"
	Msc18ExtractPowerOracle  = "5c30a92b-2e88-4e66-b0a5-fa05304136e7"
	Msc18FolkHeroOracle      = "c5126d6e-969c-4034-98e3-f8653163346f"
	Msc18HangoutOracle       = "43306a78-85b5-4df4-a018-75dd75aa8997"
	Msc18DonaldBlakeOracle   = "f966ae40-4717-4f41-9437-0e38b01a8be9"
	Msc18AlexWilderOracle    = "f3031397-c2d5-40c6-8eca-88faed00a5ea"
	Msc18DivinationOracle    = "273b339c-964b-4a18-8eb5-ceb8abcdfd9e"
	Msc18VoraciousTyphonOrcl = "f62784c0-9c67-4a05-a09b-aabf03a9390f"
)

func Msc18DivinationCast(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return CastCatalogSpell(t, g, "Divination", "Sorcery", Msc18DivinationOracle, nil)
}

func Msc18Library(me *game.Player, n int) {
	names := make([]string, n)
	for i := range names {
		names[i] = "Library Card"
	}
	SeedLibrary(me, names...)
}

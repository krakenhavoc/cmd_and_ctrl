package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_x_from_graveyard_test.go — #1948, ADR 0189's 2026-10-16
// amendment: "exile X cards from your graveyard" as a spell's additional
// cost (the X is the one announced with the cast), and a fixed count on a
// branch of an either/or cost (Soaring Stoneglider).

const (
	SkeletalScryingOracle   = "561c68d8-b727-48d8-a832-2e42d1345fac"
	HarvestPyreOracle       = "00e36265-0cd8-4084-adf5-d1ed44f0d689"
	ChillHauntingOracle     = "ebe61471-e236-4ff8-9416-151970b5efd6"
	HauntingMiseryOracle    = "8db6651b-c368-46fd-b9c2-69a4d6562246"
	SoaringStonegliderOracl = "62d18d2b-fa68-49cd-b62d-803728ddfa8b"
)

// exileXFuel puts n cards of `typeLine` into p's graveyard.
func ExileXFuel(p *game.Player, typeLine string, n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ids[i] = uuid.New()
		p.Graveyard.PushTop(game.Card{InstanceID: ids[i], Name: "Fuel", TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	}
	return ids
}

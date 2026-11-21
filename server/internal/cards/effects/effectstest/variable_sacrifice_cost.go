package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// variable_sacrifice_cost_test.go — ADR 0100 sub-PR 4: a variable
// sacrifice count as a cast's additional cost (CR 601.2b, 601.2f–h,
// 107.3a). The rules the board shows — what the announcement may name,
// what the count prices and what the resolution reads back — and one
// test per card.

const (
	ViciousBetrayalOracle         = "be1015a7-2ace-4b97-8884-202abae8401b"
	DevouringRageOracle           = "19a8c2d3-b06e-4a76-bb28-8928ce84186c"
	DevouringGreedOracle          = "efbe96b5-1f58-4818-a346-ea554813adbc"
	EliminateTheCompetitionOracle = "26e549c0-a08b-475b-9138-6dde175cdf55"
	ImmoralBargainOracle          = "c82cc78b-ffd6-4f72-881d-85913830feb2"
	DevastatingSummonsOracle      = "5eb6626a-1e0a-438e-9e95-a1e86be6489d"
	PlumbTheForbiddenOracle       = "b099fc54-cdbc-46cb-b4e4-7c2ca77b115b"
	TorgaarOracle                 = "4229140f-fa5b-4727-a16a-cbbd756d979e"
	DargoOracle                   = "a2414a2d-0fd8-4084-8762-ace48f70c853"
	RottenmouthViperOracle        = "2c75623c-59f4-4449-ab43-9d1225185ad9"
)

// vsPermanent puts a permanent onto the battlefield under `owner` and
// returns its ID.
func VsPermanent(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return EitherPermanent(g, owner, name, typeLine)
}

func VsCreatures(g *game.Game, owner uuid.UUID, n int, typeLine string) []uuid.UUID {
	out := make([]uuid.UUID, n)
	for i := range out {
		out[i] = VsPermanent(g, owner, "Fodder", typeLine)
	}
	return out
}

func VsTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

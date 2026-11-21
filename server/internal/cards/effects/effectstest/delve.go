package effectstest

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// delve_test.go — CR 702.66, ADR 0100 sub-PR 1: delve as a payment.
// The pure arithmetic (the budget, X, the any-colour fold) is tested
// next to the code in game/delve_test.go; these are the board-level
// rules — what the announcement may name, when the cards are exiled,
// and what the record keeps.

const (
	TreasureCruiseOracle   = "5b6bdf5a-2742-4851-92cd-a857a3852836"
	DigThroughTimeOracle   = "f8b17b89-26ce-4208-874a-9e1d66514640"
	MurderousCutOracle     = "c0ce7d5a-68fd-41f3-b5ba-a1a178cee9ec"
	DeathRattleOracle      = "ae327976-1026-4318-8846-f6b0cac373a7"
	BecomeImmenseOracle    = "40e5b83e-5f53-4b15-8bfd-c0c0b8355a6f"
	MagmaticSinkholeOracle = "2da7f1b7-316b-4e8e-9fa3-5880b258d601"
	SetAdriftOracle        = "f27f51e1-b881-4639-9d36-47cdf2a61117"
	TasigursCrueltyOracle  = "a232daf3-db54-4711-88b0-e5072c05f58e"
	WillOfTheNagaOracle    = "520b6637-0a9f-4dc4-846e-f1cd2b868263"
	LogicKnotOracle        = "b2da7acb-d80c-414c-9f7d-753a5d6ccad9"
	EmptyThePitsOracle     = "089d91cf-ed7a-4859-8967-cad975a5127e"
	RiteOfUndoingOracle    = "9dbfa026-e364-4111-a03a-e9b1693bc7b7"
	DeadDropOracle         = "09f3d60c-34ee-41ec-a047-fe2140b11950"
	SibsigOracle           = "423c1079-bdba-42fc-8732-59cb83ebafc3"
	DivinationDelveOracle  = "273b339c-964b-4a18-8eb5-ceb8abcdfd9e"
)

// delveFuel puts n instants into p's graveyard and returns their IDs,
// oldest first.
func DelveFuel(p *game.Player, n int) []uuid.UUID {
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		id := uuid.New()
		p.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: fmt.Sprintf("Fuel %d", i), TypeLine: "Instant",
			Owner: p.ID, Controller: p.ID,
		})
		ids = append(ids, id)
	}
	return ids
}

func CastCruise(t *testing.T, g *game.Game, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	return CastWithTapParams(t, g, "Treasure Cruise", "Sorcery", "{7}{U}", TreasureCruiseOracle, params)
}

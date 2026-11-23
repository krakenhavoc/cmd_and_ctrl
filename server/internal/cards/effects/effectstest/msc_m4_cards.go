package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mscM4TryCast puts the card in the active player's hand, advances to a main
// phase and casts it, returning the cast's error so a refusal can be tested.
func MscM4TryCast(t *testing.T, g *game.Game, name, typeLine, oracle string, modes []int, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	AdvanceToMain(t, g)
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

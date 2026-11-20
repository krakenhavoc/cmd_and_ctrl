package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castModalPaying is castPaying with modes.
func CastModalPaying(t *testing.T, g *game.Game, name, typeLine, oracleID string, modes []int, targets []game.TargetRef,
	optional []int, teamwork, blight []uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracleID,
		Owner: active.ID, Controller: active.ID})
	AdvanceToMain(t, g)
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{
		Modes:         modes,
		Targets:       targets,
		OptionalCosts: optional,
		TeamworkIDs:   teamwork,
		BlightIDs:     blight,
	})
}

func CardRefsOf(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

func PushArtifactForTest(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name,
		TypeLine: "Artifact", ManaCost: "{2}", Owner: owner, Controller: owner})
}

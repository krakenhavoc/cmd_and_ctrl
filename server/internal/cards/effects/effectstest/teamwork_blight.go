package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// teamwork_blight_test.go — #1703: teamwork (CR 702.194a) and blight
// paid as a cost (CR 701.68a), end to end through CastSpell.

const (
	CinderStrikeOracle      = "54421e69-d79e-4c2e-8ce6-96994d168835"
	RequitingHexOracle      = "eda16b31-33db-4ad2-901a-80fbe6273733"
	RepulsorBlastOracle     = "150f920c-c942-4feb-824f-79dd9b531687"
	TwDoublingSeasonOracle  = "01546b7d-a233-4176-8843-d732074dc5b6"
	TwWindingConstrictorOra = "c9404d7d-a026-4082-9fcb-1ab571a136b5"
)

// castPaying is castWithOptionalCosts for the two #1703 components:
// the optional costs announced, the teamwork taps and the blighted
// creature.
func CastPaying(t *testing.T, g *game.Game, name, typeLine, oracleID string, targets []game.TargetRef,
	optional []int, teamwork, blight []uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracleID,
		Owner: active.ID, Controller: active.ID})
	AdvanceToMain(t, g)
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{
		Targets:       targets,
		OptionalCosts: optional,
		TeamworkIDs:   teamwork,
		BlightIDs:     blight,
	})
}

func TwCard(g *game.Game, id uuid.UUID) *game.Card { return FindBattlefieldCardByID(g, id) }

func TwTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

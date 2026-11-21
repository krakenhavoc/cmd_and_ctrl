package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choose_number_cards_test.go — #1941, ADR 0129's amendment of
// 2026-10-09: the three cards the number prompt unblocked.

const (
	CnVolcanoHellion     = "12a88c08-11e3-4f51-a17c-ca06452bac6f"
	CnPhyrexianProcessor = "36c800cb-b1ca-4432-ad3c-4d8b90337f4c"
	CnNecrodominance     = "a10b3e35-8cc4-450e-9e30-0fce8df0fea4"
)

// castHellion casts Volcano Hellion as the active seat, a 6/5, and
// passes until its enters trigger asks for a target.
func CastHellion(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Volcano Hellion", TypeLine: "Creature — Hellion",
		OracleID: CnVolcanoHellion, Power: 6, Toughness: 5, Owner: active.ID, Controller: active.ID})
	AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Volcano Hellion: %v", err)
	}
	PassPriorityAroundTable(t, g)
	return id
}

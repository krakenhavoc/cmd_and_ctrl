package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// escalate_test.go — CR 702.120a, #2126: "For each mode you choose
// beyond the first as you cast this spell, you pay an additional
// [cost]." The Collective cycle (discard, mana, tap a creature) plus
// the mana-only Borrowed and Alliance cards, through the real cast
// path.

const (
	CollectiveBrutalityOracle  = "22a78443-db21-4656-b38a-e3e3186fd94b"
	CollectiveResistanceOracle = "bffdfe7b-f17a-41b6-a460-80280fe497d4"
	CollectiveDefianceOracle   = "43ddc3e8-936b-4be0-9e64-484bbbd69016"
	CollectiveEffortOracle     = "a97e760a-99c8-47ea-a875-451aca913ee7"
	BorrowedMalevolenceOracle  = "ee6e0279-9606-4877-a3a0-5d9dd557ff15"
	BorrowedHostilityOracle    = "b45fc33b-0658-4fe1-89da-54a07a12ebd4"
	BorrowedGraceOracle        = "e858d26f-db45-4c56-b40a-bbaeedbf5549"
	SavageAllianceOracle       = "57da1d3f-7469-4efa-9b82-b2d915177d89"
	BlessedAllianceOracle      = "48b67778-6616-4563-8456-62f302649768"
)

// castEscalate seeds the spell into the active seat's hand, walks to a
// main phase and casts it with `params`, returning the spell's id and
// the announce error. The hand is NOT cleared: callers set it up.
func CastEscalate(t *testing.T, g *game.Game, name, typeLine, oracle, manaCost string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Owner: active.ID, Controller: active.ID,
	})
	AdvanceToMain(t, g)
	return id, g.CastSpell(active.ID, id, params)
}

func EscPlayerRef(id uuid.UUID, occ int) game.TargetRef {
	return ModeRef(game.TargetPlayer, id, occ, 0)
}

func EscCardRef(id uuid.UUID, occ int) game.TargetRef {
	return ModeRef(game.TargetCard, id, occ, 0)
}

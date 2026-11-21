package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// life_exchange_test.go — the four "exchange life totals" cards of
// issue #1117. Every assertion is on observable state: the two life
// totals, the cards drawn, and for Tree of Perdition the post-layer
// toughness read the way the wire reads it.

const (
	AxisOfMortalityOracle  = "158cb888-a115-4d50-a82f-746622f0578e"
	MagusOfTheMirrorOracle = "f587676e-d45f-4b1d-be2c-b404d8f55fb3"
	MisterNegativeOracle   = "c8da3262-826d-4ea8-bb22-06374bebefe5"
	TreeOfPerditionOracle  = "0cdece2a-0bdf-4e6d-9ddc-4a8d58b2ec29"
)

// --- Tree of Perdition ------------------------------------------------

func LxTree(g *game.Game, owner uuid.UUID) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tree of Perdition",
		OracleID: TreeOfPerditionOracle, TypeLine: "Creature — Plant",
		Power: 0, Toughness: 13, Owner: owner, Controller: owner,
	})
}

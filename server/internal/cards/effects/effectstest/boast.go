package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// boast_test.go — #2697, CR 702.142. The engine half is game/boast.go;
// the enumerator's and the wire's halves are in internal/legal and
// internal/protocol. Everything here is a real printed card.
//
// Fearless Liberator is the smallest boast card (a mana cost and a
// token), Birgi is the limit modifier, Frenzied Raider watches the
// activation, Dragonkin Berserker prices it, Broadside Bombardiers
// reads the sacrificed permanent.

const (
	FearlessLiberatorOracle  = "b0b93253-6432-401e-9a01-94c41fb72c10"
	FrenziedRaiderOracle     = "1207881c-4c03-48e1-9ca7-a2330e8053d1"
	DragonkinBerserkerOracle = "06b77a80-4968-4385-960a-fb66bb4faa94"
	BroadsideBombardiersOrcl = "e0c00a74-bdf8-4fee-8042-42c5878e9e3c"
	VarragothOracle          = "68b41a04-8cb0-4edf-b488-a219494453ae"
	EradicatorValkyrieOracle = "06b7987d-c876-4369-9c2d-ce1d35a01097"
	TuskeriFirewalkerOracle  = "36dda03b-a3c1-4ef4-ae62-da318028a39e"
)

func PushLiberator(g *game.Game, owner uuid.UUID) uuid.UUID {
	return PushCatalogPermanent(g, owner, "Fearless Liberator",
		"Creature — Dwarf Berserker", FearlessLiberatorOracle, false)
}

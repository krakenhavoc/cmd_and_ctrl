package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// block_capacity_eot_cards_test.go — #1715's cards. The engine half
// (the capacity records, the blocksEach requirement, undo and the
// snapshot) is pinned in game/multi_block_followups_test.go.

const (
	CoastlineChimeraOracle   = "da95382c-0536-49aa-a0db-3b52926bf559"
	MountedArchersOracle     = "22517690-3ec1-49c9-951f-93eb3114423a"
	GiveNoGroundOracle       = "bae18a46-00e0-4b3e-8fdb-44cc61f30f8e"
	ValorMadeRealOracle      = "5bd1348c-51cc-4426-b995-8bb6d4a5bd2e"
	ActOfHeroismOracle       = "25edb501-11b5-4617-8a65-2eb4869cccd5"
	YareOracle               = "f0f24c6d-80fb-4e99-a74d-6ed9d349ed3d"
	BlazeOfGloryOracle       = "b330ac89-790e-4cc9-96a5-532c48252088"
	LairwatchGiantOracle     = "dac0dc4c-acf3-40de-b6b5-963ff4176d3a"
	AnuridSwarmsnapperOracle = "fd275ea7-9ccc-4148-9b33-392a612486dd"
	LuminousGuardianOracle   = "0d9a31f5-20d0-4266-95ec-38b4fa96b8aa"
)

// bcCombat is the shared table: `me` (the active seat) has n 2/2s to
// attack the defender `opp` with. The cursor is at the beginning of
// combat, where every instant here can be cast.
func BcCombat(t *testing.T, n int) (g *game.Game, me, opp *game.Player, atks []uuid.UUID) {
	t.Helper()
	g = NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	for i := 0; i < n; i++ {
		atks = append(atks, PushVanillaCreature(g, me.ID, "Bear", 2, 2))
	}
	AdvanceTo(t, g, game.StepBeginCombat)
	return g, me, opp, atks
}

// bcAttack declares every attacker at `opp` and moves to the declare
// blockers step.
func BcAttack(t *testing.T, g *game.Game, opp *game.Player, atks []uuid.UUID) {
	t.Helper()
	DeclareAttack(t, g, opp.ID, atks...)
	AdvanceTo(t, g, game.StepDeclareBlockers)
}

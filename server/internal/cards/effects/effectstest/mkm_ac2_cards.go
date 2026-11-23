package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_ac2_cards_test.go — Murders at Karlov Manor A–C, batch 2 (#4469).
// Each test plays the real catalog entry and asserts what its text says.

const (
	MkmAC2AlquistOracle      = "dd79a994-7dad-49d4-8d8f-62dbedde4a38"
	MkmAC2AudienceOracle     = "cbbb5c0c-e953-410d-88c6-8fedb719e94c"
	MkmAC2BuriedOracle       = "f1f0eab3-681b-4c2d-b590-555951920880"
	MkmAC2WitnessOracle      = "9038a457-ce01-4f6d-9e2b-5220f08b4290"
	MkmAC2CandlestickOracle  = "2f7b46c1-3d82-4fd9-9882-589315cab47b"
	MkmAC2GatewayOracle      = "42a9aeac-5b9c-4675-8c96-7c8ce7e9098c"
	MkmAC2RansackedLabOracle = "f85eea8f-5887-4921-a943-d15e0ad1e028"
	MkmAC2TrampledOracle     = "575a0dc2-206a-406a-90b3-72a92f2ea87a"
	MkmAC2ChalkOracle        = "af335eb3-f60a-44b3-a1ed-1e165f3e3605"
	MkmAC2CoercedOracle      = "3a007dbd-18fd-44d1-a305-7cc8dd6231eb"
	MkmAC2CadaverOracle      = "cc5370e0-a0e6-4958-a6f2-093a3868b575"
	MkmAC2InquiryOracle      = "52fe8101-82af-4aed-a395-54e7b7639bd2"

	MkmAC2CaseTypeLine = "Enchantment — Case"
)

// --- shared helpers ------------------------------------------------------------

// mkmAC2PushGraveyard puts a catalog card with a mana cost into a
// player's graveyard.
func MkmAC2PushGraveyard(g *game.Game, p *game.Player, name, typeLine, oracle, manaCost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
			Owner: p.ID, Controller: p.ID,
		})
	})
	return id
}

// mkmAC2AddMana puts mana into a player's pool.
func MkmAC2AddMana(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	})
}

// mkmAC2Solve solves a Case directly.
func MkmAC2Solve(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.SolveCaseForEffect(id); err != nil {
			t.Fatalf("SolveCaseForEffect: %v", err)
		}
	})
}

// mkmAC2EndStepSolves reports whether the Case is solved once the
// active player's end step trigger has resolved.
func MkmAC2EndStepSolves(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	AdvanceTo(t, g, game.StepEnd)
	PassPriorityAroundTable(t, g)
	return MkmB1Solved(g, id)
}

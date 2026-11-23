package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr1_g3_cards_test.go — ADR 0108 Delivery PR 1 (#1886, #1887),
// group 3: the modal "that creature" bullets, the "dealt damage this
// way" sweeps, a split card with fuse, and the activated and triggered
// "can't be regenerated" cards.

const (
	P1g3AgateOracle        = "7b89b7d2-c724-4d5d-9f0b-7d3302ad1168"
	P1g3SuplexOracle       = "2d1bbeda-2e81-4aaa-9494-094ec3dd6c3b"
	P1g3PineconeOracle     = "961e3023-39ea-4141-99b6-738280a2815d"
	P1g3ExpulsionOracle    = "71d178b3-e5ca-4576-83f7-8dce74758acd"
	P1g3GnashingOracle     = "188a3f61-3f7c-450d-8ee7-e12ad13b86c6"
	P1g3CrushOracle        = "4aa119f7-d411-4188-956e-547f7d14e789"
	P1g3UnderworldOracle   = "ed3d113d-f2c3-4ef3-8bdd-ed4824a22ab0"
	P1g3GallifreyOracle    = "d4424585-9564-4ec2-8267-3f5438e1f29e"
	P1g3FlamebreakOracle   = "9e608ace-2844-419b-9204-9054127557b2"
	P1g3SpikeOracle        = "c983644d-6741-4aa1-aa68-a6e680c26bb6"
	P1g3JayaOracle         = "cbbad3da-695e-4527-9678-0942f3751287"
	P1g3OrcishHealerOracle = "9b7a0c9d-1045-4fc0-b5e3-18616106da08"
	P1g3CohortOracle       = "1e3b97d2-8fda-4510-9697-f36ca9ca2ab0"
	P1g3BoOracle           = "f5dc3dbd-7eab-40f3-afe2-1e88a0c00538"
)

// p1g3ModeCard is one target announced for mode occurrence `occ`.
func P1g3ModeCard(id uuid.UUID, occ int) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id, Mode: occ}
}

// p1g3CastModal casts a modal spell with its modes and targets
// announced together (CR 601.2b).
func P1g3CastModal(t *testing.T, g *game.Game, name, typeLine, oracle string, modes []int, targets ...game.TargetRef) uuid.UUID {
	t.Helper()
	return B02cCastModalSpell(t, g, name, typeLine, oracle, modes, targets)
}

// p1g3Activate has the active seat activate row `index` of `source`
// and resolves it.
func P1g3Activate(t *testing.T, g *game.Game, source uuid.UUID, index int, params game.ActivateAbilityParams) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	if err := g.ActivateCatalogAbility(me.ID, source, index, params); err != nil {
		t.Fatalf("activate row %d: %v", index, err)
	}
	PassPriorityAroundTable(t, g)
}

// p1g3Untap untaps a permanent between activations.
func P1g3Untap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { FindBattlefieldCardForTest(g, id).Tapped = false })
}

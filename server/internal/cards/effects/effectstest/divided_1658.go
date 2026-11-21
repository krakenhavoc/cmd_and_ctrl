package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// divided_1658_test.go — #1658, the cards #1656's divided damage and
// counters (CR 601.2d) unblocked. One test per card for its main
// split, one refusal per card for a bad split (a zero share or a sum
// mismatch), and the trigger path for the three ETB creatures
// (Verdurous Gearhulk, Armament Dragon, Glint Weaver).

const (
	RollingThunderOracle        = "9c47888b-28a5-4c43-9ee4-a9059e3c367d"
	ForkedBoltOracle            = "e0548f0d-4aec-4c39-8c34-52bfa2bf7c60"
	TwinBoltOracle              = "970dc070-3668-4584-9904-2897b27bb806"
	ArcLightningOracle          = "5acc8b39-3c3e-4012-8cfd-ac3c2c4ca982"
	ElectrolyzeOracle           = "07b222d7-24f2-4994-9004-ff6672ebe161"
	MagicMissileOracle          = "e48b8d78-3f8b-4bf8-853a-76c435bc31f0"
	FlamesOfTheFirebrandOracle  = "bc32c24f-2f2a-4125-917c-60d425166640"
	ChandrasPyrohelixOracle     = "69a37af8-6bcd-42b3-b788-0a357e742cc0"
	ViolentEruptionOracle       = "842578ca-0a86-4e96-bfd5-45931488f7c1"
	PyrokinesisOracle           = "d001febc-d511-4e65-a631-91dc9415056b"
	VerdurousGearhulkOracle1658 = "3765e8bb-e70d-4503-b8dc-1e684e434c18"
	ArmamentDragonOracle        = "23c75386-8011-4ee4-97e2-eaafeb20b788"
	GlintWeaverOracle           = "eebe3fd6-afd0-4bb3-9e8f-c737497dae33"
	BlessingsOfNatureOracle     = "1e77cf90-ac51-4ac7-b123-be04aabe1688"
	SplendidAgonyOracle         = "e9958344-13d5-4f7b-acc8-904e5d99b90e"
)

// d1658Cast casts a hand card with the full CastSpellParams a divided
// spell needs (Distribution, XValue, AlternativeCost) — castCatalogSpell
// only carries Targets, and b12PlayFromHand's land-drop bookkeeping is
// noise for a spell.
func D1658Cast(t *testing.T, g *game.Game, name, typeLine, oracleID string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracleID,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// d1658CastErr is d1658Cast for a cast that is SUPPOSED to be refused:
// the announce error comes back instead of failing the test.
func D1658CastErr(t *testing.T, g *game.Game, name, typeLine, oracleID string, params game.CastSpellParams) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracleID,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g.CastSpell(active.ID, id, params)
}

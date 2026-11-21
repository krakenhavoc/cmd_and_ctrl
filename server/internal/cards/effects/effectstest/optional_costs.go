package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// optional_costs_test.go — ADR 0073 (#664), the card half.
//
// Every assertion here fails with the engine change backed out, and
// each one fails for a different reason, which is the point of having
// six cards rather than one:
//
//   - Burst Lightning: the resolution reads the paid record.
//   - Rite of Replication: the same read feeding a COUNT.
//   - Wolfbriar Elemental: a trigger on the permanent reads the count
//     AFTER entry, which is the half the stack item cannot answer.
//   - Gatekeeper of Malakir: the same read as a CR 603.4 intervening
//     if, so an unkicked one triggers nothing at all.
//   - Capsize: buyback's mana half, and the graveyard when declined.
//   - Constant Mists: buyback's NON-mana half, paid at CR 601.2h.

const (
	BurstLightningOracle  = "ac2086fe-98ee-4280-9c7c-c5c2d6548a8b"
	RiteOfReplicationOra  = "fb60739e-1dc3-481d-a056-ad72e665c680"
	WolfbriarOracle       = "2f8872fe-84dc-4cda-a253-e7503a5c96a3"
	GatekeeperOracle      = "6781f8ae-2a86-4e3d-bc43-48809c9d6c26"
	CapsizeOracle         = "77637eff-2963-4402-88f3-ca346f762fc8"
	ConstantMistsOracle   = "c850a29e-dc40-4ab6-89f1-d501a1a350d1"
	UrzasRuinousOracle    = "978e0d87-3ff2-4a73-916c-ff0dc0ab2797"
	RakdosLordOracle      = "143a269a-b9ee-48ba-bd7b-4aa46eb36778"
	GrafdiggersCageOracle = "753cb2b2-24ce-484f-a2d0-be6fd2c67ebd"
	RuleOfLawOracle       = "53e88e64-6f82-4154-a66e-6aeb0154b368"
)

// castWithOptionalCosts is castCatalogSpell plus the CR 601.2b
// announcement: the optional costs claimed, and the permanents paid
// to the non-mana half of one.
func CastWithOptionalCosts(t *testing.T, g *game.Game, name, typeLine, oracleID string,
	targets []game.TargetRef, optional []int, sacrifices []uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{
		Targets:       targets,
		OptionalCosts: optional,
		SacrificeIDs:  sacrifices,
	})
}

// findBattlefieldCardByID is a local read so these tests do not
// depend on the engine's own unexported lookup.
func FindBattlefieldCardByID(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

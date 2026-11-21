package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s23_boardwipes_test.go — S23: the mass-effect primitives and the
// cards built on them.
//
// The tests worth reading first are the SIMULTANEITY ones. Every
// other assertion here would have passed against the pre-S23
// hand-rolled loops; the Zulaport pair is the pair that would not,
// and it is the reason the sprint touched the engine at all.

const (
	WrathOfGodOracle       = "34515b16-c9a4-4f98-8c77-416a7a523407"
	SupremeVerdictOracle   = "0230de18-8d15-4cfa-9d42-7ccddd9f9570"
	FumigateOracle         = "b17ea905-0696-4e58-b564-557e87236e27"
	DeadlyTempestOracle    = "b5516bc9-ec8d-4323-8748-96c49d7d0622"
	InGarruksWakeOracle    = "a6899b94-427d-4851-a474-4087e0a0918a"
	RitualOfSootOracle     = "29e9cf1c-a6bd-4bee-9000-ac1b4e19d6b0"
	PlanarCleansingOracle  = "a98c2d81-4add-4292-bbdd-e1b69ff936d4"
	ShatterstormOracle     = "96ce2403-4607-440a-92ae-80aceb458c5d"
	CleansingNovaOracle    = "aff34f28-f707-4458-8af3-1bd5b13a6b10"
	CruxOfFateOracle       = "52a0dae4-2a95-487e-acd4-eabdb2d031e2"
	DepopulateOracle       = "4a83c2fa-f2d2-4b86-8407-4269f127936d"
	EvacuationOracle       = "fdd94383-b573-439a-8e1c-925af887c5a6"
	RiversRebukeOracle     = "c52cfb41-18f3-4e73-b5e7-d75baf74e578"
	WhelmingWaveOracle     = "e510eaaf-6497-480f-baa8-f4796b5f1086"
	ToxicDelugeOracle      = "afaef788-34d1-460b-b884-9d7ae6ddeb18"
	LanguishOracle         = "ef1a83f2-6707-41a2-b5ed-861c8e45ae07"
	MassacreWurmOracle     = "93cf50cf-0ecc-4d3e-abea-778c1ebacec4"
	ChandrasIgnitionOracle = "f61680da-606e-4d16-b0a0-361aa5210901"
	BaneOfProgressOracle   = "51f9a6cc-8eb2-44ed-a2d9-913ac514ad67"
	NevinyrralsDiskOracle  = "96230edf-568a-47dd-b877-9d92aa58fac8"
	ZulaportOracle         = "76b003e0-15af-4f22-bdf2-1ade5430964a"
)

// pushWipeCreature seeds a plain vanilla creature onto the
// battlefield with an explicit size, since most of these tests care
// about toughness thresholds.
func PushWipeCreature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		Power:      power,
		Toughness:  toughness,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// castWipe seeds a sorcery into the active seat's hand and casts it
// with explicit params, for the cards castCatalogSpell's
// targets-only signature cannot express (an X value, a mode).
func CastWipe(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracle,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

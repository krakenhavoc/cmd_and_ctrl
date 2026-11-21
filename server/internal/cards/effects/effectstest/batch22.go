package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch22_test.go — card-level coverage for the card-coverage
// roadmap's batch 22 (#384, `edhrec_rank` 2335–2435): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, land play or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b22-prefixed.

const (
	B22WhelmingWaveAlreadyOID   = "e510eaaf-6497-480f-baa8-f4796b5f1086"
	B22ConclaveMentorOracle     = "a2fe5937-212c-4e71-8d6e-f408b38100aa"
	B22RograkhOracle            = "584cee10-f18c-4633-95cc-f2e7a11841ac"
	B22ShivanGorgeOracle        = "e90a1381-c9c1-4f57-928c-5d19dc065274"
	B22DictateOfTheTwinGodsOID  = "3d960d33-623a-4415-ae00-f8cffbc15f5a"
	B22SlaughterTheStrongOracle = "7a3569a0-a55f-40c9-9588-92db35e26567"
	B22TarnishedCitadelOracle   = "66ae2562-68e9-4c77-ba0a-57f8ff37f656"
	B22SamwiseGamgeeOracle      = "7ce37c26-91ea-493f-bfc3-a890d4538bc1"
	B22DisdainfulStrokeOracle   = "11e02134-7b1a-46a4-a89e-7539dd1efada"
	B22ToxrillOracle            = "c71b2325-bde6-4364-a93b-8477ffeb25d8"
	B22CabbageMerchantOracle    = "e31808bb-fb6e-487b-a973-db44961c84ee"
	B22CauldronOfEssenceOracle  = "a2f8cde8-bf7b-4234-89f0-a95f9dc937e3"
	B22EarthshakerDreadmawOID   = "454cbc6a-b7f0-445e-842c-5db267917a18"
	B22BerserkersOnslaughtOID   = "85d2948c-1a79-418d-80c4-bc1012a4d313"
	B22FuryOracle               = "fbf9f8c5-849f-45d5-8129-5fc683c21a04"
	B22WirewoodLodgeOracle      = "1275653f-de4e-4fe9-aad8-88555fa11680"
	B22WavebreakHippocampOracle = "3405c8a9-a8d6-4b45-9b64-94141076603b"
	B22SingularityRuptureOracle = "e1976b6c-7e43-4f4a-b082-f95495a1b260"
	B22CodexShredderOracle      = "ef7b11ab-24e7-4e7c-91a7-920bade6e60b"
	B22WearDownOracle           = "27905301-333e-4cdd-90cf-188159fcf8e9"
	B22HullBreachOracle         = "2da232d8-580f-4116-b977-2c59cd21b5a4"
	B22LilianasCaressOracle     = "a4aec0d6-13fa-4709-b1a9-2f483f032744"
	B22SabotenderOracle         = "65385311-1158-4e94-892a-683997706ca8"
	B22DawnOfHopeOracle         = "d7a38484-2acc-49a6-b32d-d54dddb14d31"
	B22MycosynthWellspringOID   = "9ec43ec6-b625-4be8-8f79-3679e6657dbc"
	B22EreborFlamesmithOracle   = "6eb93545-a1a6-4447-82c5-421d8e9f023d"
	B22IorethOracle             = "af6e4c3e-0276-4f72-9a70-25868fe8bba5"
	B22AethericAmplifierOracle  = "295cd8e1-0830-46c1-9957-556afd4bcee6"
	B22GodEternalOketraOracle   = "3b03358d-f87e-4939-afc9-5ee3f044146a"
	// #1210: a declared skip for "no ability-activated event", which
	// #1184 built and #1210's watch shape put to use.
	B22HarshMentorOracle = "168336e2-d795-4b75-bf21-ce128b0dd7b0"
)

// b22Untap untaps a battlefield card through the effect API.
func B22Untap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
}

// b22CastInstantAs puts an instant in a player's hand and casts it
// from wherever the cursor is — the non-active seat's instant-speed
// cast, for a "during each opponent's turn" trigger.
func B22CastInstantAs(t *testing.T, g *game.Game, p *game.Player, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant", ManaCost: "{U}",
		Owner: p.ID, Controller: p.ID,
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// b22TryModal attempts a modal cast and returns the engine's answer.
func B22TryModal(g *game.Game, oracle string, modes []int, targets []game.TargetRef) error {
	active := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			return err
		}
	}
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Hull Breach", TypeLine: "Sorcery",
		OracleID: oracle, Owner: active.ID, Controller: active.ID})
	err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
	if err != nil {
		_, _ = active.Hand.Remove(id)
	}
	return err
}

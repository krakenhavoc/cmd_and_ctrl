package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch08_test.go — card-level coverage for the card-coverage
// roadmap's batch 08 (#301, `edhrec_rank` 902–1004): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// the earlier batch test files are reused by name; new ones are
// b08-prefixed.

const (
	B08ResculptOracle              = "814e321d-16fd-4f7b-a8d8-2b089be76f2c"
	B08WarleadersCallOracle        = "a751c07b-fc21-4854-96e7-f71abf4e86c9"
	B08ZuranOrbOracle              = "08cb8a30-9cb4-4517-bee5-8848aa60d1a2"
	B08SpiritedCompanionOracle     = "9c5f0d91-9d86-4e66-94fd-4af93ad01838"
	B08BlightedWoodlandOracle      = "02679a2e-303d-412f-87d8-0a37a8ca259c"
	B08BristlyBillOracle           = "d3b2d8a2-d3bc-448c-9cf6-6bead6010c28"
	B08OldGnawboneOracle           = "dff3f4c7-f792-4ca3-9b0a-0738e70664d9"
	B08DispelOracle                = "6d7be242-a072-40ce-b540-95880506cccd"
	B08FountainportOracle          = "94e8b0a9-44a1-4dce-8d44-78681ae638a1"
	B08SolveTheEquationOracle      = "f02682f0-26c0-4032-9df3-273b6a45d0a8"
	B08BloodfellCavesOracle        = "64e29bfc-9313-4e8c-808c-bc27f6b018a6"
	B08ThopterSpyNetworkOracle     = "49be65fd-3755-410d-b0dc-2e5861ea2552"
	B08WildernessReclamationOracle = "6f856f99-4cb4-479d-958d-964220965ed6"
	B08CoatOfArmsOracle            = "5f7f133e-58ea-41ab-b1be-be4b400fac4c"
	B08OmnathLocusOfRageOracle     = "1816eede-c5bd-49df-958f-a3af64cb2932"
	B08PhyrexianReclamationOracle  = "647ca69e-cc01-4b2b-b376-bee2a98331e8"
	B08CruelCelebrantOracle        = "3ee78cfc-0e9e-4737-a7e2-b42f94228040"
	B08ArchaeomancerOracle         = "a91a3266-cadd-47a0-9b20-160307f14c07"
	B08MindcrankOracle             = "c73c1d91-0163-49c6-832a-b9327e7a2c9b"
	B08CabarettiCourtyardOracle    = "65424bea-fd53-4f85-9757-0b91a6d40ba4"
	B08RazorkinNeedleheadOracle    = "a78f981a-bf8a-42a4-b171-d655cc2cc1a2"
	B08KrosanVergeOracle           = "d9a10971-f32b-4978-952d-fed0a5bc9e36"
	B08MoldervineReclamationOracle = "68639a3d-2192-4921-8298-c76bb0cd6b02"
	B08SythisOracle                = "0fc64fd6-f057-4056-9dca-47accb7ff036"
	B08EnchantresssPresenceOracle  = "795b096a-2bce-4588-a2c9-abc5ea40dc0c"
	B08SunkenRuinsOracle           = "e6415ffb-8b7a-41c3-bedf-0d4112b7b795"
	B08EmeriaOracle                = "cc999cf2-c99b-4911-8c52-6cc4a99fcc7b"
	B08NoxiousGearhulkOracle       = "a77b5be2-f361-4135-ba25-670a74d268ac"
	B08CutADealOracle              = "8b5cffd5-5db3-4151-9275-d91387554412"
	B08GavonyTownshipOracle        = "8a44e4e7-dfa2-427b-bbff-11c398fa60bb"
	B08AllIsDustOracle             = "14693689-d087-43b6-9c3f-63ab0648fc20"
	B08MinesOfMoriaOracle          = "583cdebe-0195-45be-bd2e-5765f07cb902"
	B08TorbranOracle               = "8c3495bf-02e7-4ad9-949d-92eb3d2b662a"
	B08ManagorgerHydraOracle       = "b3f2265b-dd65-4b74-8b74-35ee0b147617"
)

// b08CastBolt casts a Lightning Bolt at a player and settles it. With
// `red` the card carries its printed colour, as the deck importer
// stamps it; without, the fixture is colourless — which is what
// separates "a red source" from "a source".
func B08CastBolt(t *testing.T, g *game.Game, target uuid.UUID, red bool) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	card := game.Card{
		InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: LightningBoltOracle,
		Owner: active.ID, Controller: active.ID,
	}
	if red {
		card.Colors = []string{"R"}
		card.ManaCost = "{R}"
	}
	active.Hand.PushTop(card)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target}},
	}); err != nil {
		t.Fatalf("CastSpell Lightning Bolt: %v", err)
	}
	PassPriorityAroundTable(t, g)
}

// b08Untap untaps a battlefield permanent through the effect API —
// a land that has used its tap cost once, so a second activation can
// be tested in the same turn.
func B08Untap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
}

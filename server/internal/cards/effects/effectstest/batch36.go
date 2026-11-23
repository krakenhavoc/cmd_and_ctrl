package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch36_test.go — card-level coverage for the card-coverage
// roadmap's batch 36 (#399, `edhrec_rank` 3753–3852): the "no new
// machinery" group. One test per observable behaviour, driven through
// a real cast, activation or attack rather than by calling primitives.

const (
	B36SanctumOfEternityOracle  = "c7d9ff27-f1fc-42e4-a47b-d2e6d68e4035"
	B36LlanowarVisionaryOracle  = "f75ed312-3a23-4624-80c5-03980aa22d0b"
	B36AngelOfSerenityOracle    = "f4ce6078-8c7b-4f68-b324-13130c63a983"
	B36TemperedSteelOracle      = "e220138c-5fc6-487d-9fa9-f21b2fb1f12c"
	B36AkoumRefugeOracle        = "354fecd1-2371-49e3-81c6-7e47728dbb1f"
	B36AnimalSanctuaryOracle    = "f3c40943-1d7c-4ea2-b34f-8df8b6775701"
	B36HeliodsPilgrimOracle     = "0fe01c7d-f435-44c1-82f1-0ec2c47d4704"
	B36JaggedBarrensOracle      = "64ee02f1-afdb-474b-a893-31538ad7219a"
	B36WhitemaneLionOracle      = "e8d6084b-9b72-438e-a30a-851b888f3e4d"
	B36CrucibleOfFireOracle     = "0c572396-4e53-4c86-9b04-f41341dcea05"
	B36ArahboTheFirstFangOracle = "dfec65a0-6fb6-463d-aa62-e84a427db217"
	B36ReplenishOracle          = "523ae937-5535-490d-96ea-07f331b5e5ad"
	B36RepurposingBayOracle     = "1787ac2f-762d-4f3a-b7e5-12db6d3d470d"
	B36LifecreedDuoOracle       = "63686c6b-9051-4002-aa8e-da8a3021330f"
	B36LossarnachCaptainOracle  = "1251691e-3c4c-4e45-9f59-f156852f2268"
	B36TurbulentSpringsOracle   = "9aef7510-9f06-4939-8cae-f71330d1105e"
	B36MavrenFeinOracle         = "1b94a11b-21b0-4465-a520-69608f022fb4"
	B36FanaticOfMogisOracle     = "f19d06af-6caf-41d8-8d0a-d5d50bd67900"
	B36ShardingSphinxOracle     = "9ebc0144-fc45-4de7-b3d8-ef8cf2e211ae"
	B36SyggRiverCutthroatOracle = "cd4db500-0017-46c6-be94-1bf48f686b6a"
	B36BishopOfWingsOracle      = "af4684cf-f109-44be-adfb-7c551a36635e"
	B36TegwyllOracle            = "868f0a0a-ca9e-4baf-8295-6b228aa834e5"
	B36ExtinguishAllHopeOracle  = "c7ca25f3-7a22-477b-8546-4c2597d5d1ff"
	B36HealersHawkOracle        = "28a52ba1-95da-44e1-8ac5-0dc23c902394"
	B36CadiraOracle             = "fcf22321-2f55-41ca-b8e1-4792540ba3ee"
	B36ErodedCanyonOracle       = "852c6520-d148-4923-a312-05a9af821f24"
	B36SarkhansTriumphOracle    = "c4f7cdb7-aef9-4ab8-995b-4e5180d47214"
)

// b36Commander seeds a creature flagged as a commander, owned and
// controlled by `owner`, able to attack.
func B36Commander(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: owner, Controller: owner, IsCommander: true,
	})
}

// b36CastCreature casts a creature with a real power and toughness
// from the active player's hand and settles the spell and whatever
// it triggered.
func B36CastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	return id
}

// b36Lands seeds n vanilla lands under owner.
func B36Lands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		B12Permanent(g, owner, "Wastes", "Basic Land")
	}
}

// b36AcceptCommandZone answers the CR 903.9b "put it into the command
// zone instead?" prompt with "yes".
func B36AcceptCommandZone(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionalReplacement && c.Chooser == owner {
			if err := g.ResolveOptionalReplacement(c.ID, owner, true); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
			return
		}
	}
	t.Fatalf("no command-zone prompt for %s", owner)
}

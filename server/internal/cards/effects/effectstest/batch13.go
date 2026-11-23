package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch13_test.go — card-level coverage for the card-coverage
// roadmap's batch 13 (#306, `edhrec_rank` 1421–1524): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// the earlier batch test files are reused by name; new ones are
// b13-prefixed.

const (
	B13RuggedHighlandsOracle      = "6c922206-6e68-4dcd-9559-88da1074f2c4"
	B13HedronCrabOracle           = "7216f974-3c84-40ef-b904-82019900c204"
	B13CorpseKnightOracle         = "0c77b805-728d-4ae4-88a9-223f4b7b52e0"
	B13UnderworldDreamsOracle     = "967cf377-ae26-464d-85ac-8448b5a911f7"
	B13ImperiousPerfectOracle     = "3fa71348-fa4d-4f39-a451-cf1570591991"
	B13DrownInTheLochOracle       = "0f264e5b-264e-4e97-9a8d-8ae1d6a286ce"
	B13BoltwaveOracle             = "bb8cccc8-e56f-42fa-a0d7-fc889b5e1c7b"
	B13GoldveinHydraOracle        = "2b62543f-a475-457a-a96b-b5d070383d3c"
	B13InspiringVantageOracle     = "3f17c60e-923a-4392-9da8-87d9ded009b7"
	B13LongRiversPullOracle       = "f1993767-1d07-49c8-b8dc-04ec9840a999"
	B13DarkDealOracle             = "c527eb80-ccac-40b8-8377-c31121613128"
	B13KessigFlamebreatherOracle  = "eadd9559-6dcb-4b96-8c95-58abddd0e120"
	B13BeledrosWitherbloomOracle  = "90194ff1-db61-463f-b5a3-15cd85311d0e"
	B13VibrantCityscapeOracle     = "6a7f3e1f-6798-4644-b64c-7765f81f0938"
	B13WittyRoastmasterOracle     = "34edbf1b-0826-49f1-a04d-15cc8c543b22"
	B13NoMercyOracle              = "b9538d53-480b-481a-abbf-83ab17e1a45b"
	B13FellTheMightyOracle        = "1fbb2a7c-8093-4729-b8fe-cf032d99470f"
	B13ThaliaHereticCatharOracle  = "13be47e4-9da2-4f9e-baf0-db96f7777ccb"
	B13ConcordantCrossroadsOracle = "ff01b408-6d17-40a3-9efd-a1b341ec1307"
	B13WorldShaperOracle          = "3c075bb6-1831-4521-bd8d-4ed2825ae796"
	B13LaezelOracle               = "c066f921-e349-45d1-8ec3-0955d10bbf19"
	B13FarhavenElfOracle          = "4ce2357f-93e6-40ca-beca-8f4e15adc464"
	B13SunbakedCanyonOracle       = "f97fd068-b83a-4621-bf8c-cc96e880ce90"
	B13HeliodsInterventionOracle  = "e7564d66-767c-4cd9-a5f0-0f2488a4a74b"
	B13ProsperousInnkeeperOracle  = "da785227-cf8a-4d44-9e7c-fc909ea868f2"
	B13TwitchingDollOracle        = "fd6e1967-237a-41f6-bbf4-2c869f9447c8"
	B13CopperlineGorgeOracle      = "a05f641c-15c9-43dc-ae0d-1ea372fd33d5"
	B13DourPortMageOracle         = "cf58e309-00e8-438e-813e-2e1c1002db23"
	B13CrystalVeinOracle          = "616d6013-24f4-4999-9bf3-5b0764e52fa6"
	B13GratuitousViolenceOracle   = "7c340a39-4ee0-4ba1-bb66-6674f8020fda"
	B13WaveGoodbyeOracle          = "06ef46c6-00ba-40e6-b866-d0095ab83749"
	B13DranaOracle                = "89b24b5f-d837-4274-8877-8ff7dc2708ba"
	B13RadiantGroveOracle         = "32c91719-f3dd-4cc7-9e32-7d5ccf18f07c"
	B13InGarruksWakeAlreadyOID    = "a6899b94-427d-4851-a474-4087e0a0918a"
)

// b13OpponentCasts puts a card of the given type line in an
// opponent's hand and has them cast it during the active seat's
// precombat main, returning the spell's ID while it is on the stack.
func B13OpponentCasts(t *testing.T, g *game.Game, opp *game.Player, name, typeLine, oracle, manaCost string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	AdvanceToMain(t, g)
	id := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("opponent CastSpell %s: %v", name, err)
	}
	return id
}

// b13PlayAs walks the cursor to the given seat's precombat main and
// has that seat play a card from hand — a land play or a cast, the
// engine decides by type.
func B13PlayAs(t *testing.T, g *game.Game, seat int, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	AangAdvanceToMain(t, g, seat)
	p := g.Seats[seat]
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s as seat %d: %v", name, seat, err)
	}
	return id
}

// b13Tapped reads a battlefield card's tapped state.
func B13Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return B12Card(t, g, id).Tapped
}

// b13HandCard seeds a card into a player's hand and returns its ID.
func B13HandCard(p *game.Player, name, typeLine, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

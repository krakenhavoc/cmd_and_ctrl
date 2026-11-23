package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch01_test.go — card-level coverage for the card-coverage
// roadmap's batch 01 (#294, `edhrec_rank` 9–237): the "no new
// machinery" group, the two until-end-of-turn cards #314 unblocked,
// and the two mana-from-a-spell cards the AddMana primitive
// unblocked. One test per observable behaviour, driven through a
// real cast or activation rather than by calling primitives.

const (
	ArcaneDenialOracle            = "ab1cc360-b9de-48d9-9983-4dfe4a7d2a37"
	ManaDrainOracle               = "74d3277a-38e5-4732-afed-084a56148f20"
	VictimizeOracle               = "240e85d3-e495-4877-8609-4b4056c402f7"
	SpectatorSeatingOracle        = "cf6d10ed-85c3-48f2-8ba0-2960e03b408b"
	VaultOfChampionsOracle        = "ebc5ac83-08d4-4d6b-b840-0c4ba71a38ab"
	FarewellOracle                = "4eb813fd-2d5a-4b02-8193-662681ef4e7d"
	UndergrowthStadiumOracle      = "7c69f718-acc8-4851-8e5d-0cbaaa86192c"
	DreamrootCascadeOracle        = "dd8538e6-cd5f-4a88-aff5-eb5e76ce8ddb"
	TirelessProvisionerOracle     = "ab8d5f5c-1976-4f77-8ed2-8d28ee666741"
	StormcarvedCoastOracle        = "4722105b-0085-4bb8-bca1-9de0d3eb5600"
	SpireGardenOracle             = "45fe016e-1a09-410c-bbe3-4663ba06c5b7"
	BuriedRuinOracle              = "3644f316-f9a3-46c9-9b1e-747f86cf4ead"
	ImpactTremorsOracle           = "9242cd3e-1a71-4700-8182-9c1005616033"
	PhyrexianTowerOracle          = "1861e642-21d5-4232-89f3-b5557f2946c1"
	StormKilnArtistOracle         = "a145ff8c-5812-4bcb-bd16-9839dc25121d"
	StrokeOfMidnightOracle        = "9a107e48-3d50-4941-95b1-10f2b29a4245"
	RapidHybridizationOracle      = "06692cd9-ac2f-4a32-8fd1-043ba3c0fe71"
	ProfessionalFaceBreakerOracle = "04152e7a-969c-4858-841b-0a569a9fc1bf"
	BountifulPromenadeOracle      = "761cb262-f83b-4a99-9345-b773182a7671"
	FyndhornElvesOracle           = "df317532-7d36-40fd-938f-e972749c8792"
	OrnithopterOfParadiseOracle   = "3a940dfa-a026-4969-8981-ef90cdfbe9ef"
	KarplusanForestOracle         = "bd912666-f37f-4767-af6f-9e6d0fcccacf"
	TerminateOracle               = "6257c2fd-005f-41e3-8a72-af76df1eb134"
	RockfallValeOracle            = "185c70c1-8403-4ae5-b45d-3679d4ee092a"
	MirkwoodBatsOracle            = "0636b6c3-0662-420a-b30d-f0a14e7c512d"
	BrushlandOracle               = "5eb8b497-ec9a-4a89-ad29-1ec3ca82da7c"
	DarkRitualOracle              = "53f7c868-b03e-4fc2-8dcf-a75bbfa3272b"
	ReturnOfTheWildspeakerOracle  = "2b76f9e9-cd28-4eaf-8674-215c34263f96"
	BorosCharmOracle              = "2679d0dd-ba30-4a1c-b6a0-b3ac6c790496"
)

// batch01AdvanceToStepOf walks the cursor until `seat` is active and
// the turn is at `step`.
func Batch01AdvanceToStepOf(t *testing.T, g *game.Game, seat int, step game.Step) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if g.Turn.Step == step && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward %s of seat %d: %v", step, seat, err)
		}
	}
	t.Fatalf("never reached %s of seat %d", step, seat)
}

// batch01OpponentCasts puts a card in a non-active player's hand and
// casts it at instant speed (the sorcery gate only applies to
// non-instants), returning its ID so a counterspell can target it.
func Batch01OpponentCasts(t *testing.T, g *game.Game, opp *game.Player, name, oracle, manaCost string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle, ManaCost: manaCost,
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("opponent CastSpell %s: %v", name, err)
	}
	return id
}

func Batch01PoolColors(p *game.Player) []string {
	out := make([]string, 0, p.ManaPool.Total())
	for _, tok := range p.ManaPool.Units() {
		out = append(out, tok.Color)
	}
	return out
}

func Batch01GraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// --- the 2026-09 slice: the cards the September seams unblocked ----
//
// Twenty-four more of #294's hundred, written against the primitives
// that landed after the first pass: conditional alternative costs,
// until-end-of-turn grants, protection as a quality, modal triggers,
// trigger doubling, abilities from the hand, the one-shot land-drop
// grant, layer-4 dependency ordering, and the produced-mana grammar's
// per-colour amounts.

const (
	OptOracle                    = "713332c1-5bd8-400f-bfff-c1ca0697a043"
	BrainstormOracle             = "36cd2364-d113-47d1-b2c4-b088d9eb88dd"
	GrowthSpiralOracle           = "34bcc217-dd91-45a0-90d7-a94d02f1f317"
	ExploreOracle                = "b8f566ea-8283-4afa-9ac8-737e26419283"
	CropRotationOracle           = "28b46183-c62f-47b1-9fee-3ba148202cab"
	AshBarrensOracle             = "58257464-278e-45fa-8e0b-bcd9a7500bc1"
	FlawlessManeuverOracle       = "4e183439-17d2-47ff-9d99-5e22821d91e3"
	AkromasWillOracle            = "fd949f82-fc10-4e37-8aa9-6c7569fe3c55"
	JeskasWillOracle             = "0fd114c4-092b-4e28-b0dc-ef529f3bc73e"
	MysticSanctuaryOracle        = "17b60106-a4c7-410a-8ac3-ec8e74e29a7c"
	SecludedCourtyardOracle      = "79ba18fd-f184-43c1-86df-56ee18ce806c"
	NykthosOracle                = "84dc18f0-8225-4b40-a165-b10321e41769"
	ThreeTreeCityOracle          = "da3b17a2-e1e1-44e9-b9b1-ae54a92037db"
	PatchworkBannerOracle        = "4fb00dbe-1f82-4ba6-b18c-97e816d10d3a"
	YavimayaOracle               = "8dd5f5af-d2d8-4356-8617-8381081b930c"
	ScuteSwarmOracle             = "aa854d50-444c-49d9-bfb1-5476b33c1c0b"
	BlackMarketConnectionsOracle = "d2664f28-49e1-46f8-a863-b217e961a57c"
	HeraldsHornOracle            = "c02c5547-b9c9-4b2d-9d12-e87bfba8f2d2"
	RoamingThroneOracle          = "3640c29b-1534-4952-b297-619ade948431"
	TheGreatHengeOracle          = "78427103-9543-41fb-b6d4-72963fe87275"
	BoseijuWhoEnduresOracle      = "bf1341dd-41a3-49f6-87ec-63170dde4324"
	OtawaraSoaringCityOracle     = "e9b6a394-691c-425a-9307-76d8edc7375e"
	CommandBeaconOracle          = "7e8c2a18-e404-40ff-a9e0-ec3eeb6d576e"
	GemstoneCavernsOracle        = "c0adbddc-b070-4c5f-afe0-0474c72a9251"
)

// --- slice helpers ---------------------------------------------------

// castCatalogSpellWithModes is castCatalogSpell for a modal card: the
// modes are announced with the cast (CR 601.2b).
func CastCatalogSpellWithModes(t *testing.T, g *game.Game, name, typeLine, oracleID string, modes []int) uuid.UUID {
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
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

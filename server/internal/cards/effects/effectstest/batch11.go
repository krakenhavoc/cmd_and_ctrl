package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch11_test.go — card-level coverage for the card-coverage
// roadmap's batch 11 (#304, `edhrec_rank` 1214–1315): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// the earlier batch test files are reused by name; new ones are
// b11-prefixed.

const (
	B11MillikinOracle             = "fa4dffda-6f04-4d0b-829d-28a1a5794dee"
	B11GlaringFleshrakerOracle    = "0ffce5e0-6b1d-4d1a-9318-1a1a219df532"
	B11IronMyrOracle              = "6c5cbab6-ee27-46f5-97a7-df85698d1e9f"
	B11SilverMyrOracle            = "66e8f7f8-3a6d-46ba-837c-b9713ddf7f40"
	B11AbandonedAirTempleOracle   = "9575d7ce-f26d-4b90-87a3-6329e9799572"
	B11MagdaOracle                = "3d268e48-3004-4384-bf52-e63243fb5e02"
	B11ExemplarOfLightOracle      = "9ad730b5-8950-4215-9319-387e2970dd22"
	B11SecureTheWastesOracle      = "b2347910-d6c6-4681-8316-7ef27056485c"
	B11RosieCottonOracle          = "168d5711-4459-440f-8de4-aabffd47c44d"
	B11FractureOracle             = "f21d0319-0509-4ac1-b6e3-10955a26fd7a"
	B11WightOfTheReliquaryOracle  = "4507df69-6bf7-43d6-a609-c032b61835d5"
	B11KeeperOfTheAccordOracle    = "49b9a3ac-d245-431a-bf40-64d1b8dc2b91"
	B11MahadiOracle               = "1b3e841e-0f8f-467d-9983-f9b8d081a67f"
	B11BalefireDragonOracle       = "3d783beb-9ca6-4681-9276-fc3ad13b993f"
	B11MaestrosTheaterOracle      = "9464ddf2-4bcb-44f6-b945-89a132544de6"
	B11PeerIntoTheAbyssOracle     = "21fa2442-6eac-4dce-a9cc-76f0053fdb8f"
	B11HauntedMireOracle          = "b0b58a03-462c-4964-97c7-42bc777ec23e"
	B11SylvanScryingOracle        = "ee24bf27-484d-4e1c-998e-6a74e3d3f6c4"
	B11RuinCrabOracle             = "8afc00d4-a1c6-4329-af2c-a7f58a0c33e7"
	B11AmuletOfVigorOracle        = "dba16032-66c1-4ccb-9d65-d41ac550d182"
	B11NukaColaOracle             = "6cfb03e5-aca6-4fe2-a3f1-93e1f0cbf9e1"
	B11RuneScarredDemonOracle     = "14aefe99-fb60-4f1c-a71f-7ffbe94c8b13"
	B11DefenseOfTheHeartOracle    = "e7e1b166-9267-426d-897d-24903327b48d"
	B11TendershootDryadOracle     = "a336c10a-b5bd-47ff-ba2d-31e27af1e15a"
	B11GuildlessCommonsOracle     = "ee723c7c-ec9f-4ffb-8f36-cd7637eb1fae"
	B11BrokersHideoutOracle       = "bd002797-a545-4bee-88bf-b878436e7cca"
	B11BaneOfProgressAlreadyOID   = "51f9a6cc-8eb2-44ed-a2d9-913ac514ad67"
	B11VorinclexAlreadyOID        = "5a3fdf5a-bff8-4896-b288-3f43f9a72d9b"
	B11CleansingNovaAlreadyOID    = "aff34f28-f707-4458-8af3-1bd5b13a6b10"
	B11KambalAlreadyOID           = "4987c458-604a-4727-b360-170616e91e67"
	b11LightningHelixOracleForTst = "800c258a-cfc4-4a54-a667-065ea8dea69e"
)

// b11Push seeds a catalog permanent through the timestamped push so
// its statics, replacements and triggers are live.
func B11Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b11Helix casts Lightning Helix at a player for the active seat —
// three damage there, three life here.
func B11Helix(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	CastCatalogSpell(t, g, "Lightning Helix", "Instant", b11LightningHelixOracleForTst,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target}})
	PassPriorityAroundTable(t, g)
}

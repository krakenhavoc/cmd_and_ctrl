package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch31_test.go — card-level coverage for the card-coverage
// roadmap's batch 31 (#394, `edhrec_rank` 3245–3345): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, damage, land play or step
// change. Helpers from the earlier batch test files are reused by
// name; new ones are b31-prefixed.

const (
	B31GoblinInstigatorOracle     = "8b022754-6d16-470e-b754-4df6e4f4709e"
	B31BlossomingTortoiseOracle   = "5e1bf23b-7fb0-45ff-8544-fce9fa3eba00"
	B31SealOfCleansingOracle      = "a75dbe70-7e3e-446f-9a76-9fbb414f2e7c"
	B31LeoninWarleaderOracle      = "8b1351e6-165e-4ca3-96d5-4774b3176362"
	B31KinjallisSunwingOracle     = "b9243163-b726-432e-830f-86132aa7f34a"
	B31SunscorchedDesertOracle    = "256b8c23-589e-429d-9e6e-433d55079eb4"
	B31TheWarringTriadOracle      = "df0b5995-117b-4ba8-964d-ea3c592621ef"
	B31FainTheBrokerOracle        = "31b990e3-9bad-4b0a-a9d5-f5b9ed2ad0b0"
	B31SoulsMajestyOracle         = "50fd4cf3-c347-4eb9-ab15-a9b0c5ea8b0f"
	B31TillerEngineOracle         = "62f6baba-da3b-45b8-a3c1-efb75763cca8"
	B31RapaciousGuestOracle       = "54accd4a-b471-4ae5-b3b2-a5cec44023b7"
	B31NephaliaDrownyardOracle    = "6429b4ed-1845-4643-9a3d-85f7c12f2bba"
	B31TheEndstoneOracle          = "57ee0730-3ac5-4f17-ac82-bd4a6673c2bb"
	B31VampireOfTheDireMoonOracle = "5a938371-8428-48e3-85ed-6a84a39b29c6"
	B31GemstoneMineOracle         = "0c828f10-4775-492f-9224-1e2814ad2cad"
	B31FlagstonesOfTrokairOracle  = "f73979bb-91a5-4388-b70b-0cd7a4e14291"
	B31WeaponsManufacturingOracle = "e48a8160-cffe-4e00-a3e9-a59d7fc7b3a2"
	B31MagusOfTheMoonOracle       = "6800d01c-345d-4932-a2a4-8df0fb1902f2"
	// Already on main from the S26 tribal work (tribal_lords.go).
	B31GoblinKingAlreadyOracle = "d236b3fc-0d3f-4d99-875d-e32a33fe5767"
)

// b31Push seeds a permanent with a mana cost, colours and P/T on the
// battlefield with a layer timestamp and no summoning sickness.
func B31Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle, manaCost string, power, toughness int, colors ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Colors: colors, Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b31PlayLand has the active seat play a land with the given type
// line from hand — playLandFromHand with the type line as a
// parameter, for a Desert.
func B31PlayLand(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	return B13PlayAs(t, g, g.Turn.ActiveSeat, name, typeLine, oracle)
}

// b31Tapped reads a battlefield card's tapped state.
func B31Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// b31AddMana adds coloured mana to a player's pool.
func B31AddMana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

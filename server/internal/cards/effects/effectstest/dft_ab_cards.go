package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dft_ab_cards_test.go — Aetherdrift (DFT), the cards from A to B (#4464).
// Each test plays the real catalog entry and asserts what its text says.

const (
	DftABAatchikOracle      = "405c4a7b-dbb3-48a1-9205-7dbe6e2ad363"
	DftABJockeyOracle       = "c3dc6b0e-8790-4bd3-b166-63b2dc77488f"
	DftABAetherjacketOracle = "062c8e40-dea7-4df7-bcaa-f1d696fe61c3"
	DftABAirResponseOracle  = "18a22d9f-81bc-4ed5-bd83-e3f8ec5e38c8"
	DftABVendettaOracle     = "6f053f0d-c9c7-4fcd-b218-a29d1d080104"
	DftABRunnerOracle       = "0ca1f76d-ae77-490a-8b32-d30ad33afc2d"
	DftABBeastriderOracle   = "19a0d191-64c0-4e8a-8cb0-bf680cc681bb"
	DftABBestowOracle       = "58a17110-b2a7-4dbc-bba5-5ed3728ad261"
	DftABSloopOracle        = "cec10abd-2a2d-4d89-9ecc-ea50c2ac1538"
	DftABBounceOffOracle    = "d9c29e6e-5949-4599-bf74-b9539e753f76"
	DftABGearhulkOracle     = "0c7019a7-3cca-481a-9af0-abedb89c15ad"
	DftABRamblerOracle      = "21aaecc7-b9ec-4068-8e47-4dffaa12f102"
	DftABBarrageOracle      = "dbf9feb3-0751-47bd-a8dc-ed1c076cf8c7"
	DftABBrokenWingsOracle  = "5e316864-d55c-496f-8f46-773567896864"
	DftABBroodheartOracle   = "1b759831-abe8-4996-a897-3c6d6d30a1c7"
	DftABBurnerRocketOracle = "0d07eb9f-46f4-466d-acb6-609437c34f13"

	DftABVehicle = "Artifact — Vehicle"
)

// --- fixtures ------------------------------------------------------------------

// dftABPush puts a catalog permanent with printed P/T onto the battlefield
// under owner's control, past summoning sickness.
func DftABPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// dftABCast casts a catalog card from the hand with no payment and lets
// it and anything it triggers resolve, stopping at the first prompt.
func DftABCast(t *testing.T, g *game.Game, me *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	PassPriorityAroundTable(t, g)
	return c.InstanceID
}

// dftABCrew crews a Vehicle with the given creatures and resolves it.
func DftABCrew(t *testing.T, g *game.Game, me *game.Player, vehicle uuid.UUID, idx int, crewers ...uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, vehicle, idx, game.ActivateAbilityParams{CrewIDs: crewers}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	PassPriorityAroundTable(t, g)
	if c, ok := BattlefieldCard(g, vehicle); !ok || !c.IsCreature() {
		t.Fatal("the crewed Vehicle is not a creature")
	}
}

func DftABCardRef(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func DftABInGraveyard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

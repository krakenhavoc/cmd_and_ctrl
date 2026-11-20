package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_mp2_cards_test.go — Murders at Karlov Manor cards M–P, second batch
// (#4469). Each test plays the real catalog entry and asserts what its
// text says.

const (
	MkmMP2MacabreOracle     = "04633ecd-eac0-426a-83a5-8e88262fca30"
	MkmMP2SnufflerOracle    = "8f85088c-256d-4925-a4dd-aafcc2632282"
	MkmMP2MassacreOracle    = "ccb39f4a-2684-4eb7-9066-aad81ae08962"
	MkmMP2YouthsOracle      = "9acb74a9-f372-43ad-a80d-4f8d52236eb8"
	MkmMP2MelekOracle       = "23140c6f-c674-4d13-9b78-bce0ffcb981f"
	MkmMP2SpyOracle         = "87ccb151-a948-4e5a-878b-0d46c921a2b7"
	MkmMP2NoWitnessesOracle = "0639bd82-084f-47d5-b158-05dfc29efa40"
	MkmMP2OfficiousOracle   = "e57e448b-1186-4662-a6b0-01a9ce33b961"
	MkmMP2OutColdOracle     = "2d16d51b-47e0-4cdc-83dd-a805ba50bd83"
	MkmMP2EnforcerOracle    = "be8d3788-709b-4d10-861c-a6f27d4b63a6"
	MkmMP2PersuasiveOracle  = "b401179f-eded-4973-8182-df4627133886"
	MkmMP2GadaboutOracle    = "0361b9b7-9097-4aae-aa06-c63a5fb47a76"
	MkmMP2PrivateEyeOracle  = "20750317-061b-4b65-9fc8-43477fcd7a21"
	MkmMP2ProjektorOracle   = "58b27a79-ba63-45d4-84ee-423ce0d91f6b"
	MkmMP2ThoroughfareOracl = "de5b995c-9691-4555-9070-66bcbc29f955"

	MkmMP2SpyTypeLine = "Creature — Merfolk Detective"
)

// --- shared helpers ------------------------------------------------------------

// mkmMP2Grave puts a plain card into a player's graveyard.
func MkmMP2Grave(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// mkmMP2Clues lists the Clues a player controls.
func MkmMP2Clues(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Clue" && c.Controller == controller {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

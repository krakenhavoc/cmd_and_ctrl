package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// yshtola_deck_sweep_a_test.go — the first slice of the cards the
// "Y'shtola, Night's Blessed" deck request asked for (#3657): Final
// Judgment, Hypnotic Sprite // Mesmeric Glare, Murderous Rider // Swift
// End, Torrential Gearhulk, Lyse Hext, Hraesvelgr of the First Brood,
// Thancred Waters, Ardbert, Papalymo Totolymo and Tome of Legends.

const (
	YsAFinalJudgmentOracle = "69ef9308-bc25-4f3c-a5fd-1ebaed6e4842"
	YsASpriteOracle        = "1f9efad1-9f51-47aa-a5a7-9c8103435a01"
	YsARiderOracle         = "1080c5b5-6651-4c6a-93e6-099fbe389e26"
	YsAGearhulkOracle      = "a6395447-677d-4c39-8eda-2d57e527c94e"
	YsALyseOracle          = "47bce63e-8edb-43c1-b273-4b86da5c616b"
	YsAHraesvelgrOracle    = "44cbf70c-61cf-473e-901d-039956ced642"
	YsAThancredOracle      = "dc36f81e-d1b0-44a4-bf82-4cd276bc0812"
	YsAArdbertOracle       = "a9243760-b9b1-4d43-8550-3528b5571dae"
	YsAPapalymoOracle      = "334e83cd-b513-4bd4-89a2-86875667e6c4"
	YsATomeOracle          = "cada2a1c-5db2-4702-9a57-cbe7da1bf208"
)

// ysAAdventure builds an Adventure card the way deck.toGameCard does.
func YsAAdventure(owner uuid.UUID, oracle string, creature, adventure game.Face) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutAdventure,
		Owner: owner, Controller: owner, Faces: []game.Face{creature, adventure},
	}
	c.SetFace(0)
	return c
}

// findOnBattlefieldByOracle returns the first battlefield card with the
// oracle id, or uuid.Nil.
func FindOnBattlefieldByOracle(g *game.Game, oracle string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == oracle {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

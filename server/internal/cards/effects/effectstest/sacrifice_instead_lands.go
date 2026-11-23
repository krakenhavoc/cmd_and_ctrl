package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrifice_instead_lands_test.go — ADR 0098 Decision 11: "If this land
// would enter, sacrifice <n> <kind> instead. If you do, put this land
// onto the battlefield. If you don't, put it into its owner's
// graveyard."

var SacrificeInsteadLandOracleIDs = map[string]string{
	"Heart of Yavimaya":      "6c9a854c-0509-4ed4-9d94-c45b823b65e5",
	"Kjeldoran Outpost":      "8b370db5-dfb9-4ea0-9017-bae3e767b041",
	"Lake of the Dead":       "bdf476e5-1d57-4b17-b45b-d52fd75aadeb",
	"Balduvian Trading Post": "7647940e-c99c-401c-ad1d-9ec730f66b6f",
	"Soldevi Excavations":    "5baa7abe-5bdf-40ce-9a83-a93b7cae71a3",
	"Lotus Vale":             "01fc5bb3-ebd7-4ab4-8aef-2ece1e1d9b7c",
	"Scorched Ruins":         "6ee68855-c8c5-422b-88da-163c09a96416",
}

// pushSacrificeCandidate puts a basic-typed land onto the active seat's battlefield.
func PushSacrificeCandidate(g *game.Game, owner uuid.UUID, name, typeLine string, tapped bool) uuid.UUID {
	id := PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Owner: owner, Controller: owner,
	})
	if tapped {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Tapped = true
				}
			}
		})
	}
	return id
}

func SacrificeEventsFor(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventSacrifice && ev.CardID == id {
			n++
		}
	}
	return n
}

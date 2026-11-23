package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_pick_cards_test.go — the four cards #1214's three prompt
// kinds unblocked, end to end.

const (
	IntuitionOracle       = "3c9faba7-f2d3-4978-be94-020dc8003dc0"
	GiftsUngivenOracle    = "58aec411-167d-4709-8560-793eaaed62c5"
	ScapeshiftOracle      = "c235015e-a7a9-4f8d-bf4a-cf68b3847f83"
	TragicArroganceOracle = "8a29bd35-33ef-4317-9fe5-8aaff5d7d64d"
)

// seedLibrary puts named cards on a library, and returns them in the
// order a search over the library walks them.
func SeedTypedLibrary(p *game.Player, typeLine string, names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(names))
	for _, name := range names {
		id := uuid.New()
		p.Library.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine,
			Owner: p.ID, Controller: p.ID,
		})
		ids = append(ids, id)
	}
	return ids
}

// searchPromptFor returns the open search prompt for a seat, or nil.
func SearchPromptFor(g *game.Game, seat uuid.UUID) *game.PendingChoice {
	return LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, seat)
}

// pushNamedPermanent puts one permanent on the battlefield and returns
// its ID. Caller holds g.mu.
func PushNamedPermanent(g *game.Game, controller uuid.UUID, name, typeLine string) uuid.UUID {
	return PushPermanent(g, controller, game.Card{Name: name, TypeLine: typeLine})
}

// cardInZone finds a card in a zone by instance ID.
func CardInZone(z *game.Zone, id uuid.UUID) (game.Card, bool) {
	if z == nil {
		return game.Card{}, false
	}
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

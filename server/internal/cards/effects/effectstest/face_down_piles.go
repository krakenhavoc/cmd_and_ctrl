package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// face_down_piles_test.go — #2147. Sauron's Ransom and the four other
// cards that look at the top of a library and separate it into a
// face-down pile and a face-up pile.

const (
	RansomOracle             = "fe04a81b-413c-4bc8-bb83-29978da9a43d"
	FavorOracle              = "9baff93e-4ef9-404d-884b-f651a11f3742"
	AtrisOracle              = "8b2b00ab-f1c5-4057-9957-d7daac95a847"
	RiddlesOracle            = "78bed291-0ec7-466f-bc7a-9dae7d0b56ae"
	CuratorOfDestiniesOracle = "9938d178-0ce6-45c0-b317-fd5c54231579"
)

func KnownBy(g *game.Game, id, seat uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && c.KnownBy[seat]
}

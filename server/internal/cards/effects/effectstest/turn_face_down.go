package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// turn_face_down_test.go — #1209's four proof cards. One test per
// card, each asserting the half of the seam only that card reaches:
//
//	Backslide           the targeted turn, and the way back UP again
//	Cyber Conversion    the turn with no way back up (CR 708.7)
//	Ixidron             the batch, and a size that counts what it hid
//	Master of the Veil  the loop: morph → turned face up → turn one down

// morphCardOnBattlefield seats a creature whose CARD prints morph, so
// the CR 702.37e half of the seam has something to read. The morph
// cast itself is #1194's and is tested there; what matters here is
// that the DECLARATION is on the card while the permanent is face up.
func MorphCardOnBattlefield(g *game.Game, owner uuid.UUID, name, oracleID string) uuid.UUID {
	id := uuid.New()
	PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Human Wizard",
		OracleID:   oracleID,
		ManaCost:   "{1}{U}",
		Power:      1,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

func FaceDownCardForTest(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == id {
			return c
		}
	}
	t.Fatalf("permanent %s is not on the battlefield", id)
	return game.Card{}
}

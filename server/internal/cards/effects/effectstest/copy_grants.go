package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// copy_grants_test.go — the catalog half of CR 707.9a / CR 707.9b.
// Phantasmal Image and Sakashima the Impostor are the two cards
// whose except clause grants something, and between them they cover
// all three of it: an added subtype, a granted TRIGGERED ability and
// a granted ACTIVATED one. The case that matters most is the last
// one in the file — a Clone copying a Phantasmal Image gets both,
// which is what makes a grant a copiable value rather than a
// decoration on one permanent.

const (
	OraclePhantasmalImage = "bde94af8-faea-41ff-8eed-ba642eac9968"
)

// castDoomBladeAt casts the catalog's Doom Blade at `victim` from
// the active seat, which is the cheapest way to make something
// "become the target of a spell".
func CastDoomBladeAt(t *testing.T, g *game.Game, victim uuid.UUID) uuid.UUID {
	t.Helper()
	return CastCatalogSpell(t, g, "Doom Blade", "Instant", DoomBladeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
}

// findBattlefieldByID is the by-value battlefield lookup that does
// not fail the test when the card is gone — several assertions here
// are about absence.
func FindBattlefieldByID(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

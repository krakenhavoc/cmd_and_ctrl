package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emblem_targeting_cards_test.go — ADR 0109 Delivery PR 6, the card
// half: §5's emblem gates (#1899: Narset Transcendent, Dovin Baan) and
// §6's targeting restrictions (#1885: Ground Seal, Dennick, Silent
// Gravestone, Underworld Cerberus, Tomik). The bot enumerator's and the
// view's agreement is asserted in internal/legal and internal/protocol.

const (
	NarsetOracle             = "e1de0c94-0ecc-425a-9b92-27c2745d07e7"
	DovinBaanOracle          = "5b33bfbf-e4e1-43de-9c96-ccdc2916510a"
	TomikOracle              = "9895a33f-9bbd-4440-8c1a-0d401431b77f"
	GroundSealOracle         = "13f6f960-ef79-4f2c-8874-90fe9e77099e"
	SilentGravestoneOracle   = "803815c5-12be-48b3-a101-f416c1ef9b7b"
	UnderworldCerberusOracle = "a36c2b76-d595-42ba-b8c2-bb8f02639981"
	RegrowthOracle           = "e6e4a8bd-5c40-4654-8de1-0da9afed90fd"
	GideonOfTheTrialsOracle  = "a9bbaad7-c016-4908-a6a7-2c26112a6bf6"
)

// a109p6Seats is the active seat and the next one.
func A109p6Seats(g *game.Game) (*game.Player, *game.Player) {
	seat := g.Turn.ActiveSeat
	return g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
}

// a109p6LegalCards is the legal card targets of `spec` for a spell or
// ability `by` controls.
func A109p6LegalCards(g *game.Game, by uuid.UUID, spec *game.TargetSpec) []uuid.UUID {
	var out []uuid.UUID
	g.WithWriteLock(func() {
		out = g.LegalTargetsForEffect(game.TargetSource{Controller: by}, spec).Cards
	})
	return out
}

func A109p6Has(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

// a109p6Walker puts a catalog planeswalker under p at a main phase.
func A109p6Walker(t *testing.T, g *game.Game, p *game.Player, name, oracle string, loyalty int) uuid.UUID {
	t.Helper()
	ToMain(t, g)
	return PushCatalogWalker(g, p.ID, name, oracle, loyalty)
}

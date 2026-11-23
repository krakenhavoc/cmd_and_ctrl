package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_rebound_cards_test.go — the cards that GIVE a spell rebound
// (#1854, ADR 0107 §3 decision 5). The mechanic is pinned in
// game/granted_rebound_test.go; what is pinned here is that each
// catalog declaration reaches it.

const (
	CastThroughTimeOracle = "0a911cb2-caae-4378-bf09-1c5b0751dd35"
	TaigamOracle          = "e4cf3710-8600-4f95-abb2-faefdc25693d"
)

// reboundTriggersFor counts the queued rebound delayed triggers for
// the exiled card `id`.
func ReboundTriggersFor(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.Body.Key() == "rebound/cast" && dt.Params.Object.ID == id {
			n++
		}
	}
	return n
}

// castPlainSpell casts a {0} spell of `typeLine` with no rebound of its
// own from p's hand.
func CastPlainSpell(t *testing.T, g *game.Game, p *game.Player, typeLine string) uuid.UUID {
	t.Helper()
	return CastFromHandForTest(t, g, p, "Plain "+typeLine, typeLine, "{0}", "", game.CastSpellParams{})
}

// battlefieldCardPtr is a pointer to the battlefield card `id`.
func BattlefieldCardPtr(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	t.Fatalf("%s is not on the battlefield", id)
	return nil
}

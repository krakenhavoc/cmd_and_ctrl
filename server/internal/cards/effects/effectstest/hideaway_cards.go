package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hideaway_cards_test.go — ADR 0091 (#1331) through the catalog: the
// Hideaway trigger's look-and-choose, and the four proof cards' linked
// free play. The engine half (the FaceDownHidden viewer rule, the link,
// playing a face-down card out of exile) is game/hideaway_test.go.

// stackTopLibrary pushes named cards onto the TOP of p's library, the
// first name ending up on top. Returns their IDs in that order.
func StackTopLibrary(p *game.Player, names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(names))
	for i := len(names) - 1; i >= 0; i-- {
		c := game.NewCard(names[i], p.ID)
		c.TypeLine = "Sorcery"
		c.ManaCost = "{3}{R}"
		ids[i] = c.InstanceID
		p.Library.PushTop(c)
	}
	return ids
}

func HiddenCardsIn(g *game.Game) []game.Card {
	var out []game.Card
	for _, c := range g.Exile.Cards {
		if c.FaceDownKind == game.FaceDownHidden {
			out = append(out, c)
		}
	}
	return out
}

// playHiddenGrant is the free-play permission over a hidden card, or
// nil.
func PlayHiddenGrant(g *game.Game, id uuid.UUID) *game.CastPermission {
	var perm *game.CastPermission
	g.ReadSnapshot(func() { perm = g.CastPermissionOnCardByIDForEffect(id) })
	return perm
}

// enterWithHideaway puts a hideaway card onto the battlefield from
// seat 0's hand through the real entry — cast, or a land play — lets
// the hideaway trigger resolve, and answers its pick with `pick`.
func EnterWithHideaway(t *testing.T, g *game.Game, c game.Card, pick uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast / play %s: %v", c.Name, err)
	}
	PassPriorityAroundTable(t, g)
	AnswerChooseCards(t, g, me.ID, pick)
	PassPriorityAroundTable(t, g)
}

func CardOnBattlefield(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

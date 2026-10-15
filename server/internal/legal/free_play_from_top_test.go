package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #3801: "look at the top card of your library, you may cast it free, else
// it goes to your hand" (Ziatora's Envoy). The offer is a may_cast the bot
// answers like any other; accepting must leave the looked-at card a
// library-top cast the enumerator offers and CastSpell accepts, and a
// declined one must be in the hand with nothing left to offer.
func offerTopSpell(t *testing.T) (g *game.Game, seat *game.Player, top game.Card) {
	t.Helper()
	g = newTable(t)
	seat = g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	seat.Library.Cards = nil
	top = game.NewCard("Free Sorcery", seat.ID)
	top.TypeLine = "Sorcery"
	top.ManaCost = "{1}{U}"
	top.Layout = "normal"
	seat.Library.PushTop(top)
	g.WithWriteLock(func() {
		if err := g.OfferFreePlayFromTopOrHandForEffect(game.FreePlayFromTop{
			Player: seat.ID, Source: uuid.New(), SourceName: "Test Envoy", MaxManaValue: 3,
		}); err != nil {
			t.Fatalf("offer: %v", err)
		}
	})
	return g, seat, top
}

func TestEnumeratorOffersAFreeCastFromTheLookedAtTop(t *testing.T) {
	g, seat, top := offerTopSpell(t)

	// The prompt itself is an enumerated answer pair; accept it.
	var accept *legal.Move
	for _, m := range legal.EnumerateFor(g, seat.ID) {
		if strings.HasSuffix(m.Label, ": cast it") {
			m := m
			accept = &m
		}
	}
	if accept == nil {
		t.Fatalf("the may_cast offer has no accept move: %v", labels(legal.EnumerateFor(g, seat.ID)))
	}
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(accept.Type), Player: accept.Player, Caller: seat.ID, Params: accept.Params,
	}); err != nil {
		t.Fatalf("accepting: %v", err)
	}

	moves := legal.EnumerateFor(g, seat.ID)
	got := castMovesFor(moves, top.InstanceID)
	if len(got) == 0 {
		t.Fatalf("no free cast offered off the looked-at top: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

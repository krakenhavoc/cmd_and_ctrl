package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// free_cast_x_test.go — CR 107.3b for the bot enumerator (#831).
//
// The engine refuses a non-zero X on a cast that pays neither the
// spell's mana cost nor an alternative cost that includes X. A bot
// that keeps picking a move the engine refuses stalls, so the
// enumerator has to agree: for such a cast it offers X=0 or it offers
// nothing at all.

// exileCardWithGrant drops a card into the shared exile pile carrying
// `grant`, known to the whole table the way a face-up exile is.
func exileCardWithGrant(g *game.Game, p *game.Player, c game.Card, grant game.CastPermission) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner = p.ID
	c.Controller = p.ID
	c.KnownBy = make(map[uuid.UUID]bool, len(g.Seats))
	for _, s := range g.Seats {
		c.KnownBy[s.ID] = true
	}
	g.Exile.PushTop(c)
	g.GrantCastPermissionOverCardForEffect(c.InstanceID, grant)
	return c.InstanceID
}

// castMovesFor returns the cast moves an enumeration produced for one
// card, undecoded — self_cost_modifier_test.go's castsOf already owns
// the decoded shape, and this test wants the move labels too.
func castMovesFor(moves []legal.Move, source uuid.UUID) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Kind == legal.KindCast && m.Source == source {
			out = append(out, m)
		}
	}
	return out
}

// The enumerator never offers an illegal X for a free cast. Today
// that holds because it enumerates casts from hand and the command
// zone only, so a cascade hit sitting in exile under its {0} grant
// produces no cast move at all — and the assertion is written against
// the RULE rather than against that fact, so it keeps holding when
// exile joins the sources.
//
// The ordinary hand cast of the same card is in the same table to
// show the enumerator has not simply stopped announcing X.
func TestEnumeratorOffersNoIllegalXForAFreeCast(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	mana(g, seat, 6)

	stroke := game.Card{Name: "Test Stroke", TypeLine: "Sorcery", ManaCost: "{X}{U}", Layout: "normal"}
	fromHand := handCard(seat, stroke)
	free := exileCardWithGrant(g, seat, stroke, game.CastPermission{
		Player: seat.ID, UntilTurn: g.Turn.Number, Cost: "{0}", CastOnly: true,
	})

	moves := legal.EnumerateFor(g, seat.ID)

	// Paying the printed cost, X is a real choice and the enumerator
	// takes the largest affordable one.
	hand := castMovesFor(moves, fromHand)
	if len(hand) != 1 {
		t.Fatalf("hand casts of an {X} sorcery = %d, want 1: %v", len(hand), labels(moves))
	}
	if x := xValueOf(t, hand[0]); x <= 0 {
		t.Errorf("hand cast announced X=%d, want the largest affordable X", x)
	}

	// The free cast announces 0 or is not offered.
	for _, m := range castMovesFor(moves, free) {
		if x := xValueOf(t, m); x != 0 {
			t.Errorf("free cast %q announced X=%d, want 0 (CR 107.3b)", m.Label, x)
		}
	}

	// And the enumeration as a whole is still sound: every move it
	// offers, the engine accepts.
	dispatchAll(t, g, seat.ID, moves)
}

package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// free_cast_x_view_test.go — CR 107.3b on the wire (#831).
//
// The client decides locally whether to open the X picker, so the
// server has to tell it when there is nothing to ask. Two offers
// carry the answer and both are pinned here, because the failure is
// silent in opposite directions: a missing flag opens a picker whose
// value the announce gate will reject, and a flag set too eagerly
// casts a real X spell for X=0 without asking.

func TestFreeCastOffersShipXLockedAtZero(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-view-free-x"

	prevAlts := game.CatalogAlternativeCosts
	game.CatalogAlternativeCosts = func(id string) []game.AlternativeCost {
		if id != oracle {
			return nil
		}
		return []game.AlternativeCost{
			// "You may cast this spell without paying its mana cost."
			{Key: "free", Label: "Cast without paying its mana cost"},
			// A price of its own, with no X in it.
			{Key: "flat", Label: "Pay {R}", ManaCost: "{R}"},
			// A price that keeps the X, so X is still announced.
			{Key: "kicked", Label: "Pay {X}{R}", ManaCost: "{X}{R}"},
		}
	}
	t.Cleanup(func() { game.CatalogAlternativeCosts = prevAlts })

	seen := map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}
	inHand := game.NewCard("Test Stroke", me.ID)
	inHand.TypeLine = "Sorcery"
	inHand.ManaCost = "{X}{U}"
	inHand.OracleID = oracle
	inHand.KnownBy = seen
	me.Hand.PushTop(inHand)

	v := ViewOfGameFor(g, me.ID.String())
	var card *CardView
	for i := range v.Seats[0].Hand.Cards {
		if v.Seats[0].Hand.Cards[i].InstanceID == inHand.InstanceID.String() {
			card = &v.Seats[0].Hand.Cards[i]
		}
	}
	if card == nil {
		t.Fatalf("hand view missing the seeded card")
	}
	want := map[string]bool{"free": true, "flat": true, "kicked": false}
	if len(card.AlternativeCosts) != len(want) {
		t.Fatalf("offers = %+v, want %d", card.AlternativeCosts, len(want))
	}
	for _, offer := range card.AlternativeCosts {
		if offer.XLockedAtZero != want[offer.Key] {
			t.Errorf("offer %q: x_locked_at_zero = %v, want %v", offer.Key, offer.XLockedAtZero, want[offer.Key])
		}
	}
}

// The exile grant half: a cascade hit is priced {0} and locks X, an
// impulse grant with no price of its own charges the printed cost and
// does not.
func TestExileGrantShipsXLockedAtZero(t *testing.T) {
	for _, tc := range []struct {
		name     string
		grant    game.CastPermission
		manaCost string
		want     bool
	}{
		{"cascade hit", game.CastPermission{Cost: "{0}", CastOnly: true}, "{X}{U}", true},
		{"impulse exile, printed cost", game.CastPermission{}, "{X}{U}", false},
		{"airbend {2} on a spell with no X", game.CastPermission{Cost: "{2}"}, "{3}{U}", false},
		{"free cast of a spell with no X", game.CastPermission{Cost: "{0}"}, "{3}{U}", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := buildActiveGame(t)
			me := g.Seats[0]
			c := game.NewCard("Test Stroke", me.ID)
			c.TypeLine = "Sorcery"
			c.ManaCost = tc.manaCost
			c.KnownBy = map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}
			grant := tc.grant
			grant.Player = me.ID
			grant.UntilTurn = g.Turn.Number
			g.Exile.PushTop(c)
			g.GrantCastPermissionOverCardForEffect(c.InstanceID, grant)

			v := ViewOfGameFor(g, me.ID.String())
			var got *ExilePlayView
			for i := range v.Exile.Cards {
				if v.Exile.Cards[i].InstanceID == c.InstanceID.String() {
					got = v.Exile.Cards[i].ExilePlay
				}
			}
			if got == nil {
				t.Fatalf("exile view missing the grant")
			}
			if got.XLockedAtZero != tc.want {
				t.Errorf("x_locked_at_zero = %v, want %v", got.XLockedAtZero, tc.want)
			}
		})
	}
}

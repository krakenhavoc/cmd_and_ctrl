package heuristic_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// dredge_test.go — #2390. A dredge offer is priced against the draw it
// replaces: the dredged card, whether the bot can play it soon, and the
// mill at what the bot's graveyard is worth, against the mean of the
// cards it has seen from its library. And a dredge that would leave the
// library short is declined whatever it returns.

const dredgeChoiceID = "00000000-0000-4000-8000-000000002390"

// yesNoMoves are the two answers legal offers a "may" prompt.
func yesNoMoves(t *testing.T, id string) []legal.Move {
	t.Helper()
	return []legal.Move{
		choiceMove(t, 0, id, "yes", map[string]any{"apply": true}),
		choiceMove(t, 0, id, "no", map[string]any{"apply": false}),
	}
}

func withDredgeOffer(n int, source string) viewOpt {
	return withChoice(protocol.PendingChoiceView{
		ID: dredgeChoiceID, Kind: "optional_replacement",
		Chooser: seatID(0).String(), FromPlayer: seatID(0).String(), Count: 1,
		Source: source, Dredge: n,
	})
}

func withLibraryCount(n int) seatOpt {
	return func(p *protocol.PlayerView) {
		p.Library = protocol.ZoneView{Kind: "library", Owner: p.ID, Count: n}
	}
}

// lands puts n untapped lands on seat 0's battlefield, cards 100+.
func lands(n int) []protocol.CardView {
	out := make([]protocol.CardView, n)
	for i := range out {
		out[i] = land(cardID(100+i), 0)
	}
	return out
}

func dredgeDecision(t *testing.T, n int, source string, seat []seatOpt, board []protocol.CardView) string {
	t.Helper()
	v := newView([]protocol.PlayerView{newSeat(0, seat...), newSeat(1)},
		withTurn(5, 0, "draw"), withBattlefield(board...), withDredgeOffer(n, source))
	in := input(0, v, yesNoMoves(t, dredgeChoiceID)...)
	return chose(t, in, decide(t, heuristic.New(), in))
}

// stinkweedImp is a dredge 5 creature worth having back.
func stinkweedImp(id string) protocol.CardView {
	return creature(id, 0, "Stinkweed Imp", 1, 2, keywords("flying", "deathtouch"), func(c *protocol.CardView) {
		c.TypeLine, c.ManaCost = "Creature — Imp", "{2}{B}"
	})
}

// graveTroll is a dredge 6 body the bot cannot price above a draw: a
// 0/0 it cannot cast for a while.
func graveTroll(id string) protocol.CardView {
	return creature(id, 0, "Golgari Grave-Troll", 0, 0, keywords("trample"), func(c *protocol.CardView) {
		c.TypeLine, c.ManaCost = "Creature — Troll Skeleton", "{4}{G}"
	})
}

func TestDredgeTakesACardWorthMoreThanADraw(t *testing.T) {
	imp := stinkweedImp(cardID(9))
	got := dredgeDecision(t, 5, imp.InstanceID,
		[]seatOpt{withGraveyard(imp), withHand(land(cardID(20), 0), creature(cardID(21), 0, "Bear", 2, 2))},
		lands(3))
	if got != "yes" {
		t.Errorf("Stinkweed Imp back for a draw: chose %q, want yes", got)
	}
}

func TestDredgeDeclinesACardWorthLessThanADraw(t *testing.T) {
	troll := graveTroll(cardID(9))
	hand := withHand(
		creature(cardID(21), 0, "Bear A", 3, 3, func(c *protocol.CardView) { c.ManaCost = "{1}{G}" }),
		creature(cardID(22), 0, "Bear B", 3, 3, func(c *protocol.CardView) { c.ManaCost = "{1}{G}" }),
		creature(cardID(23), 0, "Bear C", 3, 3, func(c *protocol.CardView) { c.ManaCost = "{1}{G}" }),
	)
	got := dredgeDecision(t, 6, troll.InstanceID, []seatOpt{withGraveyard(troll), hand}, lands(1))
	if got != "no" {
		t.Errorf("a 0/0 it cannot cast, from a library of 3/3s: chose %q, want no", got)
	}
}

// TestDredgeCountsGraveyardSynergy — the same troll is worth dredging
// once the graveyard holds cards the bot can cast from it: the mill is
// priced at what the graveyard is worth.
func TestDredgeCountsGraveyardSynergy(t *testing.T) {
	troll := graveTroll(cardID(9))
	hand := withHand(
		creature(cardID(21), 0, "Bear A", 3, 3, func(c *protocol.CardView) { c.ManaCost = "{1}{G}" }),
		creature(cardID(22), 0, "Bear B", 3, 3, func(c *protocol.CardView) { c.ManaCost = "{1}{G}" }),
		creature(cardID(23), 0, "Bear C", 3, 3, func(c *protocol.CardView) { c.ManaCost = "{1}{G}" }),
	)
	gy := withGraveyard(troll,
		graveyardCardView(cardID(30), 0, "Flashback A", "Sorcery", "{2}{B}", castable()),
		graveyardCardView(cardID(31), 0, "Flashback B", "Sorcery", "{2}{B}", castable()),
	)
	if got := dredgeDecision(t, 6, troll.InstanceID, []seatOpt{gy, hand}, lands(1)); got != "yes" {
		t.Errorf("milling six into a graveyard it casts from: chose %q, want yes", got)
	}
}

func TestDredgeDeclinesWhenTheLibraryIsShort(t *testing.T) {
	imp := stinkweedImp(cardID(9))
	hand := withHand(land(cardID(20), 0), creature(cardID(21), 0, "Bear", 2, 2))
	for _, tc := range []struct {
		library int
		want    string
	}{
		{14, "no"},  // 9 left after milling five: under the floor of ten
		{15, "yes"}, // 10 left: at the floor
		{60, "yes"},
	} {
		got := dredgeDecision(t, 5, imp.InstanceID, []seatOpt{withGraveyard(imp), hand, withLibraryCount(tc.library)}, lands(3))
		if got != tc.want {
			t.Errorf("library %d, dredge 5: chose %q, want %s", tc.library, got, tc.want)
		}
	}
}

// TestDredgePrefersACardItCanPlaySoon — one card, two boards: with the
// mana to cast it next turn the bot dredges it back; three lands short
// of it, it would rather draw.
func TestDredgePrefersACardItCanPlaySoon(t *testing.T) {
	ogre := creature(cardID(9), 0, "Ogre", 3, 3, func(c *protocol.CardView) { c.ManaCost = "{3}{R}" })
	hand := withHand(
		creature(cardID(21), 0, "Bear A", 4, 4),
		creature(cardID(22), 0, "Bear B", 4, 4),
		creature(cardID(23), 0, "Bear C", 4, 4),
		creature(cardID(24), 0, "Bear D", 4, 4),
	)
	if got := dredgeDecision(t, 3, ogre.InstanceID, []seatOpt{withGraveyard(ogre), hand}, lands(3)); got != "yes" {
		t.Errorf("castable next turn: chose %q, want yes", got)
	}
	if got := dredgeDecision(t, 3, ogre.InstanceID, []seatOpt{withGraveyard(ogre), hand}, lands(1)); got != "no" {
		t.Errorf("three mana short: chose %q, want no", got)
	}
}

// TestDredgeGrantReturnsALand — The Necrobloom's grant names no card;
// the bot prices it as its best land card in the graveyard, which it
// wants while it is short of lands and does not once it has plenty.
func TestDredgeGrantReturnsALand(t *testing.T) {
	gyLand := land(cardID(9), 0, func(c *protocol.CardView) { c.Name = "Forest" })
	hand := withHand(
		creature(cardID(21), 0, "Bear A", 3, 3, func(c *protocol.CardView) { c.ManaCost = "{1}{G}" }),
		creature(cardID(22), 0, "Bear B", 3, 3, func(c *protocol.CardView) { c.ManaCost = "{1}{G}" }),
	)
	if got := dredgeDecision(t, 2, "", []seatOpt{withGraveyard(gyLand), hand}, lands(2)); got != "yes" {
		t.Errorf("two lands, none in hand: chose %q, want yes", got)
	}
	if got := dredgeDecision(t, 2, "", []seatOpt{withGraveyard(gyLand), hand}, lands(6)); got != "no" {
		t.Errorf("six lands: chose %q, want no", got)
	}
}

// BenchmarkDredgeDecision prices a dredge offer on a late-game board —
// a full hand, forty permanents, a forty-card graveyard — which is the
// widest the draw estimate and the mill price ever read. The heuristic
// tier's bar is a 50ms p99 per decision.
func BenchmarkDredgeDecision(b *testing.B) {
	imp := stinkweedImp(cardID(9))
	gy := []protocol.CardView{imp}
	for i := range 39 {
		gy = append(gy, creature(cardID(300+i), 0, "Dead Bear", 2, 2))
	}
	var hand []protocol.CardView
	for i := range 7 {
		hand = append(hand, creature(cardID(200+i), 0, "Bear", 3, 3))
	}
	board := lands(20)
	for i := range 20 {
		board = append(board, creature(cardID(400+i), 0, "Bear", 2, 2, keywords("flying")))
	}
	v := newView([]protocol.PlayerView{newSeat(0, withGraveyard(gy...), withHand(hand...)), newSeat(1)},
		withTurn(15, 0, "draw"), withBattlefield(board...), withDredgeOffer(5, imp.InstanceID))
	var moves []legal.Move
	for _, apply := range []bool{true, false} {
		params, err := json.Marshal(map[string]any{"choice_id": dredgeChoiceID, "apply": apply})
		if err != nil {
			b.Fatal(err)
		}
		moves = append(moves, legal.Move{Type: legal.TypeResolveChoice, Player: seatID(0), Kind: legal.KindChoice, Params: params})
	}
	in := input(0, v, moves...)
	p := heuristic.New()
	b.ResetTimer()
	for range b.N {
		if _, err := p.Decide(context.Background(), in); err != nil {
			b.Fatal(err)
		}
	}
}

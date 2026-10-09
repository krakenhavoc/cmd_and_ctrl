package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// costed_option_pick_test.go — #2854's enumerator half: an option_pick
// option that costs mana is offered only when the chooser can pay it
// now, carries its price as MoveCost.Mana, and is accepted by the
// engine (dispatchAll pays it on a clone).

func costedPickTable(t *testing.T, mountains int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newTable(t)
	opp := g.Seats[1]
	for i := 0; i < mountains; i++ {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Mountain", TypeLine: "Basic Land — Mountain",
			Owner: opp.ID, Controller: opp.ID})
	}
	g.WithWriteLock(func() {
		g.QueueOptionPickForEffect(game.OptionPickPrompt{
			Chooser:  opp.ID,
			Question: "Winter's Chill",
			Options: []game.ChoiceOption{
				{Label: "Pay nothing"},
				{Label: "Pay {1}", ManaCost: "{1}"},
				{Label: "Pay {2}", ManaCost: "{2}"},
			},
			Then: func(*game.Game, int) error { return nil },
		})
	})
	return g, opp.ID
}

func TestCostedOptionPickOffersWhatTheChooserCanPayAndPricesIt(t *testing.T) {
	g, opp := costedPickTable(t, 2)
	moves := legal.EnumerateFor(g, opp)
	if len(moves) != 3 {
		t.Fatalf("enumerated %v, want all three answers", labels(moves))
	}
	for _, m := range moves {
		want := ""
		switch m.Label {
		case "Winter's Chill: Pay {1}":
			want = "{1}"
		case "Winter's Chill: Pay {2}":
			want = "{2}"
		case "Winter's Chill: Pay nothing":
			if !m.AlwaysLegal {
				t.Error("paying nothing is the always-legal answer")
			}
		}
		got := ""
		if m.Cost != nil {
			got = m.Cost.Mana
		}
		if got != want {
			t.Errorf("%q costs %q on the wire, want %q", m.Label, got, want)
		}
	}
	dispatchAll(t, g, opp, moves)
}

func TestCostedOptionPickDropsWhatTheBoardCanNoLongerPay(t *testing.T) {
	g, opp := costedPickTable(t, 2)
	// The question went up with two Mountains; one is tapped before the
	// answer. The {2} answer would be refused, so it is not offered.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Controller == opp {
			g.Battlefield.Cards[i].Tapped = true
			break
		}
	}
	moves := legal.EnumerateFor(g, opp)
	if len(moves) != 2 || hasLabel(moves, "Winter's Chill: Pay {2}") {
		t.Fatalf("enumerated %v, want paying nothing and {1}", labels(moves))
	}
	dispatchAll(t, g, opp, moves)
}

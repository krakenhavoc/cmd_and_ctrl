package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// phyrexian_grant_payments_view_test.go — ADR 0131 §2 (#2531), PR 2: the
// two payments PR 1 left on mana alone carry `phyrexian_symbols` /
// `phyrexian_granted` too, so the client can offer the life answer.

// A mana ability's own mana component carries the same pair.
func TestManaAbilityViewCountsTheGrant(t *testing.T) {
	withLifeForManaView(t)
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	seedKrrikView(g, owner)
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Test Filter", TypeLine: "Land", Owner: owner, Controller: owner,
		ManaAbilities: []game.ManaAbilityShape{
			{TapCost: true, ManaCost: "{B}", Produced: "{B}{B}", Label: "{B}, {T}: Add {B}{B}"},
			{TapCost: true, ManaCost: "{1}", Produced: "{C}", Label: "{1}, {T}: Add {C}"},
			{TapCost: true, Produced: "{B}", Label: "{T}: Add {B}"},
			{TapCost: true, ManaCost: "{B/P}", Produced: "{B}", Label: "{B/P}, {T}: Add {B}"},
		},
	})
	var c CardView
	for _, v := range ViewOfGame(g).Battlefield.Cards {
		if v.InstanceID == id.String() {
			c = v
		}
	}
	if len(c.ManaAbilities) != 4 {
		t.Fatalf("got %d mana abilities, want 4", len(c.ManaAbilities))
	}
	for i, want := range [][2]int{{1, 1}, {0, 0}, {0, 0}, {1, 0}} {
		if a := c.ManaAbilities[i]; a.PhyrexianSymbols != want[0] || a.PhyrexianGranted != want[1] {
			t.Errorf("ability %d (%s): symbols %d / granted %d, want %d / %d",
				i, a.Label, a.PhyrexianSymbols, a.PhyrexianGranted, want[0], want[1])
		}
	}
}

func queuePayUnlessView(t *testing.T, g *game.Game, payer uuid.UUID, cost string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.QueuePayUnlessForEffect(payer, uuid.New(), cost, "Pay "+cost+"?",
			func(*game.Game) error { return nil }); err != nil {
			t.Fatalf("QueuePayUnlessForEffect: %v", err)
		}
	})
}

// A pay_unless prompt counts them for its CHOOSER, not for whoever holds
// the grant elsewhere on the board.
func TestPayUnlessViewCountsTheGrant(t *testing.T) {
	withLifeForManaView(t)
	g := buildActiveGame(t)
	payer := g.Seats[1]
	seedKrrikView(g, payer.ID)
	queuePayUnlessView(t, g, payer.ID, "{1}{B}")
	v := ViewOfGameFor(g, payer.ID.String())
	if len(v.PendingChoices) != 1 {
		t.Fatalf("%d pending choices, want 1", len(v.PendingChoices))
	}
	if c := v.PendingChoices[0]; c.PhyrexianSymbols != 1 || c.PhyrexianGranted != 1 {
		t.Errorf("symbols %d / granted %d, want 1 / 1", c.PhyrexianSymbols, c.PhyrexianGranted)
	}

	// The grant on the other seat's side leaves the chooser none.
	g2 := buildActiveGame(t)
	seedKrrikView(g2, g2.Seats[0].ID)
	queuePayUnlessView(t, g2, g2.Seats[1].ID, "{1}{B}")
	c := ViewOfGameFor(g2, g2.Seats[1].ID.String()).PendingChoices[0]
	if c.PhyrexianSymbols != 0 || c.PhyrexianGranted != 0 {
		t.Errorf("opponent's K'rrik: symbols %d / granted %d, want 0 / 0", c.PhyrexianSymbols, c.PhyrexianGranted)
	}

	// And a printed {B/P} counts with no grant at all.
	g3 := buildActiveGame(t)
	queuePayUnlessView(t, g3, g3.Seats[1].ID, "{B/P}")
	c = ViewOfGameFor(g3, g3.Seats[1].ID.String()).PendingChoices[0]
	if c.PhyrexianSymbols != 1 || c.PhyrexianGranted != 0 {
		t.Errorf("printed {B/P}: symbols %d / granted %d, want 1 / 0", c.PhyrexianSymbols, c.PhyrexianGranted)
	}
}

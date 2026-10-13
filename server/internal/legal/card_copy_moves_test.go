package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// card_copy_moves_test.go — the enumerator half of CR 707.12 (ADR 0066's
// 2026-10-13 amendment, #3258). A copy of a card is offered through an
// ordinary stored cast permission, so the bot sees exactly the casts the
// engine accepts: every copy, to its holder only, and a copy that has
// ceased is no longer a move. dispatchAll proves each offered move is
// accepted.

// copyFixtures puts n free instants in p's graveyard and offers a copy of
// each, returning the copies' IDs.
func copyFixtures(t *testing.T, g *game.Game, p *game.Player, n int) []uuid.UUID {
	t.Helper()
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		id := uuid.New()
		p.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: "Copy Fixture", TypeLine: "Instant", ManaCost: "{R}",
			Owner: p.ID, Controller: p.ID,
		})
		ids = append(ids, id)
	}
	var copies []uuid.UUID
	g.WithWriteLock(func() {
		copies = g.CopyCardsAndOfferCastsForEffect(game.CardCopyOffer{
			Player: p.ID, SourceName: "Fixture", Cost: "{0}",
			Timing: game.TimingFlash, Lapse: game.LapseStaysInExile,
		}, ids)
	})
	if len(copies) != n {
		t.Fatalf("made %d copies, want %d", len(copies), n)
	}
	return copies
}

func TestEnumeratorOffersEveryCardCopyToItsHolderOnly(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	copies := copyFixtures(t, g, seat, 3)

	moves := legal.EnumerateFor(g, seat.ID)
	for _, id := range copies {
		if got := castMovesFor(moves, id); len(got) == 0 {
			t.Errorf("copy %v is not offered: %v", id, labels(moves))
		}
	}
	dispatchAll(t, g, seat.ID, moves)

	for _, p := range g.Seats {
		if p.ID == seat.ID {
			continue
		}
		for _, id := range copies {
			if got := castMovesFor(legal.EnumerateFor(g, p.ID), id); len(got) != 0 {
				t.Errorf("an opponent is offered a cast of the copy: %v", got)
			}
		}
	}
}

func TestEnumeratorOffersNoCastOfACopyThatCeased(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	copies := copyFixtures(t, g, seat, 1)

	// The holder passes: the window closes and the copy ceases to exist.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if got := castMovesFor(legal.EnumerateFor(g, seat.ID), copies[0]); len(got) != 0 {
		t.Errorf("a ceased copy is still offered: %v", got)
	}
}

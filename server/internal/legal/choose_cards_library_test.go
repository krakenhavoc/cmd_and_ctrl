package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// choose_cards_library_test.go — #745. The library-to-battlefield
// cards (Genesis Wave, Ureni, Coiling Oracle's kin) are the first to
// ask a choose_cards prompt about LIBRARY cards. A prompt kind the
// enumerator cannot answer wedges a bot seat (#499 / #544), and the
// Zone re-check is the part of the resolver this candidate zone
// exercises for the first time — so every offered answer is dispatched
// against a clone, continuation included.
func TestChooseCardsOverLibraryCardsIsAnswerable(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	lib := me.Library.Cards
	if len(lib) < 3 {
		t.Fatalf("setup: library has %d cards", len(lib))
	}
	for i := 0; i < 3; i++ {
		// Make them permanents so the continuation's move accepts them.
		lib[len(lib)-1-i].TypeLine = "Artifact"
	}
	g.WithWriteLock(func() {
		ids := g.LookAtTopOfLibraryForEffect(me.ID, 3)
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  me.ID,
			Question: "Ureni — you may put a card from among them onto the battlefield",
			Cards:    ids,
			Min:      0,
			Max:      1,
			Zone:     game.ZoneLibrary,
			Then: func(g *game.Game, picked []uuid.UUID) error {
				if len(picked) == 0 {
					return nil
				}
				_, err := g.PutFromLibraryOntoBattlefieldForEffect(picked[0], game.LibraryEntryOptions{})
				return err
			},
		})
	})

	moves := legal.EnumerateFor(g, me.ID)
	// "Choose nothing" plus one answer per card.
	if len(moves) != 4 {
		t.Fatalf("enumerated %d answers, want 4: %v", len(moves), labels(moves))
	}
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("a seat owing a choice was offered %q", m.Label)
		}
	}
	dispatchAll(t, g, me.ID, moves)
}

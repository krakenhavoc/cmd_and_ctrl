package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// scry_test.go — CR 701.22 scry: Viscera Seer, Preordain.
//
// Three things here are worth more than "the cards moved":
//
//   - "Scry 2, THEN draw a card" really is ordered. A draw that runs
//     before the player answers takes one of the cards they were still
//     deciding about, and leaves the prompt unanswerable because that
//     card is no longer in the library to put back. That was a real bug
//     in the first draft of this code.
//   - Scry is LOOK AT, not reveal. Only the scrying player may see the
//     cards; anything else hands the table the top of a library.
//   - Nothing moves until the answer arrives, so an unanswered scry
//     leaves the library exactly as it was.

const (
	VisceraSeerOracle = "f82a4e85-526d-4456-b700-7760043a31be"
	PreordainOracle   = "ac641490-ca14-48d7-8cc4-b69ce984befa"
)

// seedLibrary puts named cards on TOP of the player's library so that
// names[0] is the next card drawn.
//
// Note it uses PushTop directly rather than pushLibraryCardForTest,
// which pushes to the BOTTOM — newCatalogGame already seeds a library
// of filler cards, so a bottom push leaves the filler on top and every
// scry assertion here reads the wrong card.
func SeedLibrary(p *game.Player, names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(names))
	for i := range names {
		ids[i] = uuid.New()
	}
	// Last push wins the top slot, so walk backwards.
	for i := len(names) - 1; i >= 0; i-- {
		p.Library.PushTop(game.Card{
			InstanceID: ids[i],
			Name:       names[i],
			TypeLine:   "Sorcery",
			Owner:      p.ID,
			Controller: p.ID,
		})
	}
	return ids
}

// libraryTopNames reads the top n library card names, top-first.
func LibraryTopNames(p *game.Player, n int) []string {
	out := make([]string, 0, n)
	for i := 0; i < n && i < p.Library.Size(); i++ {
		out = append(out, p.Library.Cards[p.Library.Size()-1-i].Name)
	}
	return out
}

func ScryChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceScry && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

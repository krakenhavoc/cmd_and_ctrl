package legal_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #3955, "any number of Eldrazi cards": besides singles and pairs, the
// whole suggestion list is an answer, which a pair-capped enumeration
// would never reach.
func TestAnyNumberCollectionPickOffersTheWholeSuggestionList(t *testing.T) {
	installLegalCollection(t)
	g := newSideboardTable(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(game.OutsidePickPrompt{
			Player: me.ID, Question: "Spawnsire of Ulamog", Min: 0, Max: game.OutsideAnyNumber,
			Destination: game.ChooseCastFreeFromOutside,
		})
	})
	moves := outsideChoiceMoves(legal.EnumerateFor(g, me.ID))
	all := false
	for _, m := range moves {
		if strings.Contains(m.Label, "Alpha Sorcery") && strings.Contains(m.Label, "Beta Sorcery") &&
			strings.Contains(m.Label, "Gamma Creature") {
			all = true
		}
	}
	if !all {
		t.Errorf("no move takes all three cards:\n%s", labelsOf(moves))
	}
	dispatchAll(t, g, me.ID, moves)
}

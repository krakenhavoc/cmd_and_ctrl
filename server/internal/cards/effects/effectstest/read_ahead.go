package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// read_ahead_test.go — read ahead (CR 702.155, #2123), through the real
// read-ahead Sagas of the catalog.
//
//	CR 702.155a  Chapter abilities of this Saga can't trigger the turn it
//	             entered the battlefield unless it has exactly the number
//	             of lore counters on it specified in the chapter symbol.
//	CR 702.155b  As this Saga enters, choose a number between one and
//	             this Saga's final chapter number; it enters with that
//	             many lore counters.

const (
	CrueltyOfGixOracle    = "2b5c62af-af5f-4beb-9688-19604ca46918"
	LoveSongOracle        = "9b87b4cd-4fb9-4bb3-b4c0-de1fcc6f44a2"
	ElderDragonWarOracle  = "4945f03b-ff99-4923-b62d-3cdb60275ef1"
	ReadAheadSagaTypeLine = "Enchantment — Saga"
)

func ReadAheadPrompt(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceEntryReadAhead {
			return c
		}
	}
	return nil
}

// castReadAheadSaga casts a read-ahead Saga from the active seat and
// passes priority until its entry asks.
func CastReadAheadSaga(t *testing.T, g *game.Game, name, oracle string) uuid.UUID {
	t.Helper()
	id := CastCatalogSpell(t, g, name, ReadAheadSagaTypeLine, oracle, nil)
	PassPriorityAroundTable(t, g)
	if ReadAheadPrompt(g) == nil {
		t.Fatalf("%s resolved without asking for its starting chapter", name)
	}
	return id
}

// chooseChapter answers the open read ahead prompt with chapter n.
func ChooseChapter(t *testing.T, g *game.Game, n int) {
	t.Helper()
	c := ReadAheadPrompt(g)
	if c == nil {
		t.Fatal("no entry_read_ahead prompt is open")
	}
	if err := g.ResolveEntryReadAhead(c.ID, c.Chooser, n-1); err != nil {
		t.Fatalf("ResolveEntryReadAhead(chapter %d): %v", n, err)
	}
}

// chaptersFired lists the chapter numbers the Saga has triggered, in
// the order they fired.
func ChaptersFired(g *game.Game, saga uuid.UUID) []int {
	var out []int
	for _, ev := range g.Events {
		if ev.Kind == game.EventSagaChapter && ev.CardID == saga {
			out = append(out, ev.Amount)
		}
	}
	return out
}

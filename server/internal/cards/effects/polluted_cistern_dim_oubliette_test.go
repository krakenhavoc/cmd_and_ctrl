package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const pollutedCisternOracle = "cfcf34ba-b412-447f-b200-d841371c0a1c"

// roomsBStackLibrary puts named cards on top of p's library, the first
// listed on top.
func roomsBStackLibrary(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
		ids[i] = c.InstanceID
	}
	return ids
}

// Polluted Cistern counts the distinct card types among one mill's
// cards, once per batch; it does nothing while locked.
func TestPollutedCisternCountsTypesAmongMilledCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := roomsBCard(me.ID, pollutedCisternOracle, "Polluted Cistern", "{1}{B}", "Dim Oubliette", "{4}{B}")
	first := roomsBStackLibrary(me,
		game.Card{Name: "Milled Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Milled Golem", TypeLine: "Artifact Creature — Golem", Power: 1, Toughness: 1},
		game.Card{Name: "Milled Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Milled Bear 2", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
	)

	roomsBCast(t, g, me, c, 1) // Dim Oubliette enters; Cistern stays locked
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	roomsBSettle(t, g, me, first[0])
	for _, s := range g.Seats {
		if s.Life != lives[s.ID] {
			t.Fatalf("a locked Polluted Cistern cost %s %d life", s.Name, lives[s.ID]-s.Life)
		}
	}

	roomsBUnlock(t, g, me, c, game.DoorLeft)
	roomsBSettle(t, g, me)
	roomsBStackLibrary(me,
		game.Card{Name: "Milled Forest B", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Milled Golem B", TypeLine: "Artifact Creature — Golem", Power: 1, Toughness: 1},
		game.Card{Name: "Milled Bear B", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
	)
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: c.InstanceID})
		if err := (MillCards{Player: me.ID, N: 3}).Apply(ctx); err != nil {
			t.Fatal(err)
		}
	})
	roomsBSettle(t, g, me)
	for _, s := range g.Seats {
		want := lives[s.ID]
		if s.ID != me.ID {
			want -= 3 // land, artifact, creature
		}
		if s.Life != want {
			t.Errorf("%s life = %d, want %d", s.Name, s.Life, want)
		}
	}
}

// Dim Oubliette mills three, then returns a creature card chosen from
// the whole graveyard — the milled ones included.
func TestDimOubliettePicksFromTheWholeGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	old := pushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	c := roomsBCard(me.ID, pollutedCisternOracle, "Polluted Cistern", "{1}{B}", "Dim Oubliette", "{4}{B}")
	ids := roomsBStackLibrary(me,
		game.Card{Name: "Milled Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Milled Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Milled Island", TypeLine: "Basic Land — Island"},
	)
	libBefore := me.Library.Size()

	roomsBCast(t, g, me, c, 1)
	roomsBSettle(t, g, me, ids[1])
	if me.Library.Size() != libBefore-3 {
		t.Fatalf("library %d -> %d, want 3 milled", libBefore, me.Library.Size())
	}
	if !onBattlefield(g, ids[1]) {
		t.Fatal("the chosen milled creature did not come back")
	}
	if !me.Graveyard.Contains(old) || !me.Graveyard.Contains(ids[0]) {
		t.Fatal("other graveyard cards moved")
	}
}

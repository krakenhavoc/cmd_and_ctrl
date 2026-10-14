package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// dack_fayden_helping_hand_test.go — #3522. Dack Fayden, Helping Hand:
// reveal until X creature cards (X = opponents), goad them for the rest of
// the game, and hand each to a DIFFERENT opponent.

const dackFaydenOracle = "3b21abb0-9a58-4e62-9650-4b04ce2bfb12"

func castDack(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Dack Fayden, Helping Hand", "Legendary Creature — Human Advisor", dackFaydenOracle, nil)
	passPriorityAroundTable(t, g)
}

func battlefieldIDByName(g *game.Game, name string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func controllerByName(t *testing.T, g *game.Game, name string) uuid.UUID {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c.Controller
		}
	}
	t.Fatalf("%s is not on the battlefield", name)
	return uuid.Nil
}

// optionLabels lists the opponents a prompt offers.
func optionLabels(c *game.PendingChoice) []string {
	var out []string
	for _, o := range c.PickOptions {
		out = append(out, o.Label)
	}
	return out
}

func TestDackFaydenGoadsAndHandsEachCreatureToADifferentOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []*game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p)
		}
	}
	setLibraryTopFirst(me,
		libraryLand("Forest A"), libraryCreature("Hit 1"), libraryCreature("Hit 2"),
		libraryLand("Forest B"), libraryCreature("Hit 3"), libraryCreature("Spare"))
	castDack(t, g)

	// Three opponents, so three creatures; the fourth stays in the library.
	for _, name := range []string{"Hit 1", "Hit 2", "Hit 3"} {
		if battlefieldIDByName(g, name) == uuid.Nil {
			t.Fatalf("%s should have entered", name)
		}
	}
	if battlefieldIDByName(g, "Spare") != uuid.Nil {
		t.Error("the reveal went past the third creature card")
	}
	for _, name := range []string{"Hit 1", "Hit 2", "Hit 3"} {
		c, _ := g.LookupCardForEffect(battlefieldIDByName(g, name))
		if !c.IsGoadedBy(me.ID) {
			t.Errorf("%s is not goaded by Dack's controller", name)
		}
	}

	// First permanent: all three opponents offered.
	first := latestOptionPickFor(g, me.ID)
	if first == nil || len(first.PickOptions) != 3 {
		t.Fatalf("first prompt = %+v, want a pick among the three opponents", first)
	}
	answerChoosePlayer(t, g, me.ID, opps[1])
	// Second: the opponent already chosen is not offered.
	second := latestOptionPickFor(g, me.ID)
	if second == nil {
		t.Fatal("no second prompt")
	}
	for _, l := range optionLabels(second) {
		if l == opps[1].Name {
			t.Errorf("the second prompt offers %s again: %v", l, optionLabels(second))
		}
	}
	if len(second.PickOptions) != 2 {
		t.Fatalf("second prompt offers %v, want the two opponents left", optionLabels(second))
	}
	answerChoosePlayer(t, g, me.ID, opps[0])
	// Third: only one opponent is left, so it is given without a prompt.
	if c := latestOptionPickFor(g, me.ID); c != nil {
		t.Fatalf("a third prompt was queued with one opponent left: %v", optionLabels(c))
	}

	want := map[string]uuid.UUID{"Hit 1": opps[1].ID, "Hit 2": opps[0].ID, "Hit 3": opps[2].ID}
	for name, owner := range want {
		if got := controllerByName(t, g, name); got != owner {
			t.Errorf("%s is controlled by %v, want %v", name, got, owner)
		}
	}
	// The goad names the goader, not the new controller, and is still there.
	for name := range want {
		c, _ := g.LookupCardForEffect(battlefieldIDByName(g, name))
		if got := c.Goaders(); len(got) != 1 || got[0] != me.ID {
			t.Errorf("%s goaders = %v, want [Dack's controller]", name, got)
		}
		if !c.Goads[0].ForRestOfGame {
			t.Errorf("%s is goaded only until the goader's next turn, want the rest of the game", name)
		}
	}
}

// With fewer creature cards than opponents the library is revealed to its
// end and only the creatures it has are handed out.
func TestDackFaydenWithTooFewCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	setLibraryTopFirst(me, libraryLand("Forest A"), libraryCreature("Only Hit"), libraryLand("Forest B"))
	castDack(t, g)
	c := latestOptionPickFor(g, me.ID)
	if c == nil || len(c.PickOptions) != 3 {
		t.Fatalf("prompt = %+v, want a pick among three opponents for the one creature", c)
	}
	answerChoosePlayer(t, g, me.ID, g.Seats[(g.Turn.ActiveSeat+2)%4])
	if got := controllerByName(t, g, "Only Hit"); got != g.Seats[(g.Turn.ActiveSeat+2)%4].ID {
		t.Errorf("Only Hit controlled by %v", got)
	}
	if me.Library.Size() != 2 {
		t.Errorf("library has %d cards, want the 2 lands", me.Library.Size())
	}
}

// Every prompt in the chain is in the chooser's move list, the engine
// accepts each, and a bot taking the first offer each time finishes the
// chain with every permanent going to a different opponent.
func TestDackFaydenPromptsAreEnumerated(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	setLibraryTopFirst(me, libraryCreature("Hit 1"), libraryCreature("Hit 2"), libraryCreature("Hit 3"))
	castDack(t, g)
	seen := map[uuid.UUID]bool{}
	for step := 0; step < 3; step++ {
		moves := resolveChoiceMoves(legal.EnumerateFor(g, me.ID))
		want := 3 - step
		if want == 1 {
			if len(moves) != 0 {
				t.Fatalf("step %d: a forced hand-out was left as a prompt: %v", step, moves)
			}
			break
		}
		if len(moves) != want {
			t.Fatalf("step %d: %d moves enumerated, want %d: %v", step, len(moves), want, moves)
		}
		dispatchEveryMove(t, g, me.ID, moves)
		m := moves[0]
		if err := actions.Dispatch(g, actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: me.ID, Params: m.Params}); err != nil {
			t.Fatalf("step %d: %v", step, err)
		}
	}
	for _, name := range []string{"Hit 1", "Hit 2", "Hit 3"} {
		c := controllerByName(t, g, name)
		if c == me.ID {
			t.Errorf("%s stayed with Dack's controller", name)
		}
		if seen[c] {
			t.Errorf("two permanents went to the same opponent (%s)", name)
		}
		seen[c] = true
	}
}

// Two players: X is 1, there is one opponent, so the one creature goes to
// them without a question.
func TestDackFaydenAtTwoSeats(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// Eliminate two opponents: X is the live opponent count.
	for _, p := range g.Seats {
		if p.ID != me.ID && p.Seat != g.Seats[(g.Turn.ActiveSeat+1)%4].Seat {
			p.Eliminated = true
		}
	}
	other := g.Seats[(g.Turn.ActiveSeat+1)%4]
	setLibraryTopFirst(me, libraryCreature("Hit 1"), libraryCreature("Hit 2"))
	castDack(t, g)
	if c := latestOptionPickFor(g, me.ID); c != nil {
		t.Fatalf("a prompt was queued with one opponent: %v", optionLabels(c))
	}
	if got := controllerByName(t, g, "Hit 1"); got != other.ID {
		t.Errorf("Hit 1 controlled by %v, want the one opponent", got)
	}
	if battlefieldIDByName(g, "Hit 2") != uuid.Nil {
		t.Error("X is 1 at two seats but a second creature entered")
	}
}

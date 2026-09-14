// myriad_wedge_test.go is a REPRODUCTION, not a fix. It fails today.
//
// The wedge, in one sentence: internal/legal enumerates search-library
// answers that the engine's own Validate hook will reject, the
// heuristic deterministically picks one of them, and a seat owing a
// choice is offered nothing else — so there is no pass for
// Runner.step's forced-pass rescue to force, and the bot sleeps
// forever holding the table.
//
// The three pieces, each correct on its own:
//
//   - effects/myriad_landscape.go asks for "up to two basic lands that
//     share a land type" and enforces the pair constraint through
//     SearchLibrarySpec.Validate, because no per-card predicate can
//     express a property of the PAIR.
//   - legal/choices.go's PendingChoiceSearchLibrary branch enumerates
//     combinations(SearchCards, 1, SearchMax, MaxExpansionPerSource).
//     It never sees Validate — the hook lives on the unexported
//     searchResume frame — so it offers mismatched pairs.
//   - game/search_choice.go's ResolveSearchLibrary rejects such a pair
//     with ErrInvalidParam and deliberately does NOT dequeue the
//     choice, so the client "can try again rather than losing the
//     prompt". A human retries differently. A deterministic bot does
//     not.
//
// The trigger is the deck's BASIC COUNT, not the game state.
// MaxExpansionPerSource is 12 and combinations() fills that budget
// with 1-card picks before it reaches 2-card picks, so:
//
//   - 12 or more matching basics in the library: the budget is spent
//     on singles and NO pair is ever offered. No wedge — but Myriad
//     Landscape silently fetches one land instead of two, which is a
//     play bug of its own.
//   - 11 or fewer: one or two pairs squeeze in, and they are the
//     FIRST pairs in library order. The heuristic sums cardValue, so
//     two lands always beat one and it takes a pair. If those two
//     cards are different basic types the seat is dead, and there is
//     no third pair to fall back on.
//
// Real Commander manabases run few basics. The curated simic-ramp
// deck ships 7 Forest and 5 Island in 99 cards, so the pool is under
// the cap from the opening hand and the wedge lands on the FIRST
// Myriad Landscape activation of the game — turn 2, full library.
// That is exactly what the field replay
// (aca910f0-63c3-4a0f-bef4-5a03aef44704, 11 candidates: 6 Forest +
// 5 Island) shows. An earlier draft of this file blamed a thinned
// late-game library; that was wrong, and the field evidence says the
// triggering condition is much broader.
package aiseat_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

const oracleMyriadLandscape = "2549bc57-9ffb-4053-9f10-f2a5f792b845"

// myriadDeck is a two-colour deck with Myriad Landscape in it.
// forests/islands are how many of each basic it runs; the rest is
// padding the fetch's predicate does not match. 99 cards, like the
// curated bot decks.
func myriadDeck(owner uuid.UUID, forests, islands int) []game.Card {
	deck := []game.Card{}
	cmdr := game.NewCommander("Commander Bear", owner)
	cmdr.TypeLine = "Legendary Creature — Bear"
	cmdr.ManaCost = "{2}{G}"
	cmdr.Power, cmdr.Toughness = 3, 3
	deck = append(deck, cmdr)
	for i := 0; i < 4; i++ {
		c := game.NewCard("Myriad Landscape", owner)
		c.TypeLine = "Land"
		c.OracleID = oracleMyriadLandscape
		deck = append(deck, c)
	}
	for i := 0; i < forests; i++ {
		c := game.NewCard("Forest", owner)
		c.TypeLine = "Basic Land — Forest"
		deck = append(deck, c)
	}
	for i := 0; i < islands; i++ {
		c := game.NewCard("Island", owner)
		c.TypeLine = "Basic Land — Island"
		deck = append(deck, c)
	}
	for i := 0; i < 94-forests-islands; i++ {
		c := game.NewCard("Bear", owner)
		c.TypeLine = "Creature — Bear"
		c.ManaCost = "{1}{G}"
		c.Power, c.Toughness = 2, 2
		deck = append(deck, c)
	}
	return deck
}

// findInLibrary returns the first card in p's library with the given name.
func findInLibrary(g *game.Game, p *game.Player, name string) uuid.UUID {
	for _, c := range p.Library.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func TestMyriadLandscapeSearchWedgesABotSeat(t *testing.T) {
	// simic-ramp, the deck the field wedge was on, ships 7 Forest and
	// 5 Island in 99 cards. That is the case that matters; the others
	// bracket the expansion cap.
	for _, d := range []struct {
		name             string
		forests, islands int
	}{
		{"simic-ramp_7F_5I", 7, 5},
		{"6F_6I", 6, 6},
		{"20F_20I", 20, 20},
	} {
		t.Run(d.name, func(t *testing.T) {
			myriadWedgeRun(t, d.forests, d.islands)
		})
	}
}

func myriadWedgeRun(t *testing.T, forests, islands int) {
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("Bot%d", i), myriadDeck(uuid.Nil, forests, islands)); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(7, 8))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	me := g.Seats[0]
	seat := me.ID

	lib := game.ZoneRef{Kind: game.ZoneLibrary, Owner: seat}
	bf := game.ZoneRef{Kind: game.ZoneBattlefield}

	landscape := findInLibrary(g, me, "Myriad Landscape")
	if landscape == uuid.Nil {
		t.Fatal("no Myriad Landscape in the library")
	}
	if err := g.MoveCardByID(lib, bf, landscape); err != nil {
		t.Fatalf("move landscape: %v", err)
	}
	// Myriad Landscape's printed "enters tapped" fires on the way in;
	// untap it so the ability's {T} is payable this turn.
	if err := g.TapCard(landscape, false); err != nil {
		t.Fatalf("untap landscape: %v", err)
	}
	// Two untapped Forests to pay the {2}.
	for i := 0; i < 2; i++ {
		id := findInLibrary(g, me, "Forest")
		if err := g.MoveCardByID(lib, bf, id); err != nil {
			t.Fatalf("move forest: %v", err)
		}
	}

	// The bot's own move list has to contain the activation, or the
	// rest of this test is measuring something else.
	moves := legal.EnumerateFor(g, seat)
	act := -1
	for i, m := range moves {
		if m.Type == legal.TypeActivateAbility && m.Source == landscape {
			act = i
			break
		}
	}
	if act < 0 {
		for _, m := range moves {
			t.Logf("move: %s %s", m.Type, m.Label)
		}
		t.Fatalf("the enumerator never offered Myriad Landscape's ability (%d moves); turn=%+v", len(moves), g.Turn)
	}
	t.Logf("activating %q", moves[act].Label)
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(moves[act].Type), Player: seat, Caller: seat, Params: moves[act].Params,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}

	// Resolve the ability: pass priority until the search prompt opens.
	for i := 0; i < 20 && searchChoice(g, seat) == nil; i++ {
		holder := g.Seats[g.Turn.PriorityHolder].ID
		if err := actions.Dispatch(g, actions.Action{
			Type: actions.TypePassPriority, Player: holder, Caller: holder,
		}); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	ch := searchChoice(g, seat)
	if ch == nil {
		t.Fatal("the search prompt never opened")
	}

	// --- the wedge ------------------------------------------------
	// Run the runner's own loop by hand: enumerate, decide, dispatch.
	// A healthy table clears the choice within a move or two. A wedged
	// one rejects the same deterministic answer forever, and because a
	// seat owing a choice is offered ONLY that choice's answers, there
	// is no pass to fall back on.
	answers := legal.EnumerateFor(g, seat)
	pairs := 0
	for _, m := range answers {
		if i := strings.Index(m.Label, ": take "); i >= 0 {
			if len(strings.Fields(m.Label[i+len(": take "):])) == 2 {
				pairs++
			}
		}
	}
	t.Logf("search prompt open: max=%d candidates=%d answers=%d two-card-picks=%d",
		ch.SearchMax, len(ch.SearchCards), len(answers), pairs)
	p := heuristic.New()
	rejects := map[uuid.UUID]int{}
	startTurn := g.Turn.Number
	for step := 1; step <= 400; step++ {
		if g.Turn.Number >= startTurn+2 {
			t.Logf("step %d: the table advanced two turns past the fetch; no wedge", step)
			return
		}
		// Whoever owes something acts. Choices first, then priority.
		var actor uuid.UUID
		for _, s := range g.Seats {
			if len(legal.EnumerateFor(g, s.ID)) > 0 {
				actor = s.ID
				break
			}
		}
		if actor == uuid.Nil {
			t.Fatalf("WEDGE at step %d: NO seat has a legal move. turn=%+v stack=%d choices=%s",
				step, g.Turn, len(g.Stack.Cards), describeChoices(g))
		}
		moves := legal.EnumerateFor(g, actor)
		in := aiseat.Input{View: protocol.ViewOfGameFor(g, actor.String()), Seat: actor, Moves: moves}
		d, err := p.Decide(context.Background(), in)
		if err != nil {
			t.Fatalf("Decide: %v", err)
		}
		if d.Index == aiseat.Decline {
			if pi := aiseat.PassIndex(moves); pi >= 0 {
				d.Index = pi
			} else {
				t.Fatalf("WEDGE at step %d: the policy declines and there is no pass. choices=%s", step, describeChoices(g))
			}
		}
		chosen := moves[d.Index]
		derr := actions.Dispatch(g, actions.Action{
			Type: actions.Type(chosen.Type), Player: actor, Caller: actor, Params: chosen.Params,
		})
		if derr == nil {
			rejects[actor] = 0
			continue
		}
		rejects[actor]++
		t.Logf("step %d: engine rejected %q for %s: %v", step, chosen.Label, actor, derr)
		if aiseat.PassIndex(moves) < 0 && rejects[actor] >= 3 {
			t.Fatalf("WEDGE at step %d: %q is rejected every time (%d in a row) and no pass is on offer — runner.step's forced-pass rescue has nothing to force. choices=%s",
				step, chosen.Label, rejects[actor], describeChoices(g))
		}
	}
	t.Fatalf("WEDGE: 400 bot steps without finishing the turn. turn=%+v stack=%d choices=%s", g.Turn, len(g.Stack.Cards), describeChoices(g))
}

func describeChoices(g *game.Game) string {
	out := ""
	for _, c := range g.PendingChoices {
		if c == nil {
			continue
		}
		out += fmt.Sprintf("[kind=%s chooser=%s reason=%q max=%d cands=%d] ",
			c.Kind, c.Chooser, c.Reason, c.SearchMax, len(c.SearchCards))
	}
	if out == "" {
		return "(none)"
	}
	return out
}

func searchChoice(g *game.Game, seat uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == seat && c.Kind == game.PendingChoiceSearchLibrary {
			return c
		}
	}
	return nil
}

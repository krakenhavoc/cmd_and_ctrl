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
// Why it needs a thinned library: MaxExpansionPerSource is 12, and
// combinations() fills that budget with 1-card picks before it ever
// reaches 2-card picks. With a full Commander mana base the bot is
// only ever offered single basics, every one of which is legal. Once
// the library is down to roughly a dozen basics — a ramp deck on turn
// 8+ — pairs start being offered, the heuristic prefers a pair
// (cardValue is summed, so two lands beat one), and the first pair in
// lexicographic order is a mismatch about half the time in a
// two-colour deck.
package aiseat_test

import (
	"context"
	"fmt"
	"math/rand/v2"
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
// basics is how many of EACH basic type it runs; the rest is padding
// that the fetch's predicate does not match.
func myriadDeck(owner uuid.UUID, basics int) []game.Card {
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
	for i := 0; i < basics; i++ {
		c := game.NewCard("Forest", owner)
		c.TypeLine = "Basic Land — Forest"
		deck = append(deck, c)
	}
	for i := 0; i < basics; i++ {
		c := game.NewCard("Island", owner)
		c.TypeLine = "Basic Land — Island"
		deck = append(deck, c)
	}
	for i := 0; i < 40-2*basics; i++ {
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
	for _, basics := range []int{3, 4, 5, 6, 15} {
		t.Run(fmt.Sprintf("basics=%d", basics), func(t *testing.T) {
			myriadWedgeRun(t, basics)
		})
	}
}

func myriadWedgeRun(t *testing.T, basics int) {
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("Bot%d", i), myriadDeck(uuid.Nil, basics)); err != nil {
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
	t.Logf("search prompt open: max=%d candidates=%d", ch.SearchMax, len(ch.SearchCards))
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

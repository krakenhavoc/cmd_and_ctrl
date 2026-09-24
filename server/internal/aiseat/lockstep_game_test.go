package aiseat_test

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// lockstep_game_test.go pins #1409's harness half: under the lockstep
// schedule a seed is a GAME, not a deal. Before it, four runner
// goroutines interleaved differently on every run — one binary played
// seed 107 to a winner on turn 13 eight times out of nine and to the
// turn budget on the ninth — so a red nightly could not be replayed.

// TestLockstepSeedReplaysMoveForMove plays one four-seat heuristic
// table twice from the same seed and asserts the two move logs are
// identical, line for line.
func TestLockstepSeedReplaysMoveForMove(t *testing.T) {
	requireGameTests(t)
	const (
		seed       = 107
		turnBudget = 50
		wall       = 120 * time.Second
	)
	a := playLockstepGame(t, seed, heuristicSeats(4), turnBudget, wall)
	b := playLockstepGame(t, seed, heuristicSeats(4), turnBudget, wall)
	if len(a.moves) == 0 {
		t.Fatal("the first run recorded no moves; the log measured nothing")
	}
	for i := 0; i < len(a.moves) && i < len(b.moves); i++ {
		if a.moves[i] != b.moves[i] {
			t.Fatalf("seed %d diverged at move %d of %d/%d:\n  run 1: %s\n  run 2: %s",
				seed, i, len(a.moves), len(b.moves), a.moves[i], b.moves[i])
		}
	}
	if len(a.moves) != len(b.moves) {
		t.Fatalf("seed %d: run 1 made %d moves, run 2 made %d", seed, len(a.moves), len(b.moves))
	}
	if a.turns != b.turns || a.winner != b.winner {
		t.Fatalf("seed %d: same moves, different outcome (turns %d/%d, winner %d/%d)", seed, a.turns, b.turns, a.winner, b.winner)
	}
	t.Logf("seed %d replayed exactly: %d moves, turn %d, winner seat %d, lives %v", seed, len(a.moves), a.turns, a.winner, a.lives)
}

// --- the even mirror (#1409's policy half) ------------------------

// mirrorRoom is the board #1409 describes, dealt into a real heads-up
// game: both seats at 6 life behind the same untapped creatures — six
// Drakes (3/3 flying), five Wurms (7/7 trample), six Ogres, five Bears —
// and twelve Mountains, with every evasion class matched. One more
// Drake goes to seat edgeSeat: the one thing a draw hands one side of a
// mirror.
//
// The libraries are Mountains and nothing else. That is what makes the
// test say one thing: nobody draws another creature, nobody draws a
// Bolt (two to the face finish a player on 6, and a test a burn spell
// can pass proves nothing about attacking), and nobody decks out
// inside the budget. The only way the game ends is the side with the
// edge turning it into damage.
func mirrorRoom(t *testing.T, seed uint64, edgeSeat int) *ws.Room {
	t.Helper()
	return mirrorRoomWith(t, seed, edgeSeat, mirrorSpec{edge: "Drake", life: 6, board: [4]int{6, 5, 6, 5}})
}

// mirrorSpec is the shape of a mirror table.
type mirrorSpec struct {
	// edge names the one creature the edge seat has one more of.
	edge string
	// life is both players' starting life.
	life int
	// board is each seat's Drakes, Wurms, Ogres and Bears, before the
	// edge.
	board [4]int
}

// mirrorRoomWith is mirrorRoom with the table's shape as a parameter.
func mirrorRoomWith(t *testing.T, seed uint64, edgeSeat int, spec mirrorSpec) *ws.Room {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		cmdr := game.NewCommander("Commander Bear", uuid.Nil)
		cmdr.TypeLine = "Legendary Creature — Bear"
		cmdr.ManaCost = "{2}{R}"
		cmdr.Power, cmdr.Toughness = 3, 3
		deck := []game.Card{cmdr}
		for j := 0; j < 64; j++ {
			c := game.NewCard("Mountain", uuid.Nil)
			c.TypeLine = "Basic Land — Mountain"
			deck = append(deck, c)
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Bot%d", i), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(seed, seed+1))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for i, p := range g.Seats {
		p.Life = spec.life
		proto := map[string]game.Card{}
		for _, c := range battleDeck(p.ID) {
			if _, ok := proto[c.Name]; !ok {
				proto[c.Name] = c
			}
		}
		for _, n := range []struct {
			name  string
			count int
		}{{"Mountain", 12}, {"Drake", spec.board[0]}, {"Wurm", spec.board[1]}, {"Ogre", spec.board[2]}, {"Bear", spec.board[3]}} {
			if i == edgeSeat && n.name == spec.edge {
				n.count++
			}
			for j := 0; j < n.count; j++ {
				c := proto[n.name]
				c.InstanceID = uuid.New()
				c.Owner, c.Controller = p.ID, p.ID
				// A permanent is public; one pushed past the zone-move
				// path has to be told so, or every seat's view redacts
				// it to a nameless 0/0.
				c.AddKnowersAll(seatIDs(g))
				g.Battlefield.PushTop(c)
			}
		}
	}
	return ws.NewRoom(g, testLogger(), "")
}

// TestHeuristicMirrorResolvesInsideTheTurnBudget plays the mirror with
// the smallest edge there is — one Drake — once for each seat, so the
// edge is cashed both on the play and on the draw.
//
// Before the two-turn race the heuristic sat there to the turn budget.
// All-in put one Drake through for 3 into a player on 6, so lethalPush
// said no; every single attack was blocked at a loss, so attackValue
// said no; and nothing else was ever going to arrive. The race sees
// what a player sees: the spare Drake is 3 now and 3 next turn, and
// nothing the defender has can get past what stays home. The defender
// trades Drake for Drake instead of taking it, which is progress too —
// the edge survives every trade until the last Drake is unblockable.
func TestHeuristicMirrorResolvesInsideTheTurnBudget(t *testing.T) {
	requireGameTests(t)
	const (
		turnBudget = 25
		wall       = 120 * time.Second
	)
	const seed = 1409
	for _, edge := range []int{0, 1} {
		t.Run(fmt.Sprintf("edge=seat%d", edge), func(t *testing.T) {
			res := playGameWith(t, mirrorRoom(t, seed, edge), seed, heuristicSeats(2), turnBudget, wall, true)
			total := res.totals()
			t.Logf("edge seat %d: state=%s turns=%d winner=%d lives=%v applied=%d in %v",
				edge, res.state, res.turns, res.winner, res.lives, total.Applied, res.elapsed)
			if res.state != game.StateEnded {
				t.Fatalf("the mirror did not resolve inside %d turns (state %s, lives %v); the board it stopped on:%s",
					turnBudget, res.state, res.lives, res.board)
			}
			if res.winner != edge {
				t.Errorf("seat %d won; the seat with the extra Drake is seat %d (lives %v)", res.winner, edge, res.lives)
			}
			assertNoEnumeratorBugs(t, res)
		})
	}
}

// TestHeuristicWurmEdgeResolvesInsideTheTurnBudget is #1504: the same
// mirror at seed 1409 with the edge a Wurm (7/7 trample) instead of a
// Drake, on two scaled-down boards.
//
//   - Wurms and Bears, both players on 6. Every Bear in front of a Wurm
//     lets 5 over, so swinging everything is lethal whatever the
//     defender does: lethalPush.
//   - Wurms and Ogres, both players on 8. The all-in is not lethal (two
//     Ogres on the spare Wurm absorb all of it), so the edge has to be
//     cashed over turns: the two-turn race.
//
// Before the attacker's estimates counted trample overflow, a blocked
// Wurm connected for nothing, the spare Wurm was invisible, and all
// four games stopped at the turn budget (51 turns, lives [6 6] and
// [8 4]).
//
//   - The full mirror (#1527): six Drakes, five Wurms against six, six
//     Ogres and five Bears, both players on 6. The Drakes trade off in
//     the first race, and what is left has no two-turn kill while a
//     chumped Wurm counts as dead: the one Ogre in front of the spare
//     Wurm holds it for a turn and nothing comes of it after. Counting
//     the Wurm alive into NEXT — their Wurm is busy with my other one,
//     and no single creature they have spare can kill it — is what
//     cashes the edge. Before
//     #1527 both games stopped at the turn budget with lives [6 6].
func TestHeuristicWurmEdgeResolvesInsideTheTurnBudget(t *testing.T) {
	requireGameTests(t)
	const (
		turnBudget = 25
		wall       = 120 * time.Second
		seed       = 1409
	)
	for _, spec := range []mirrorSpec{
		{edge: "Wurm", life: 6, board: [4]int{0, 5, 0, 5}},
		{edge: "Wurm", life: 8, board: [4]int{0, 5, 6, 0}},
		{edge: "Wurm", life: 6, board: [4]int{6, 5, 6, 5}},
	} {
		for _, edge := range []int{0, 1} {
			t.Run(fmt.Sprintf("life=%d/board=%v/edge=seat%d", spec.life, spec.board, edge), func(t *testing.T) {
				res := playGameWith(t, mirrorRoomWith(t, seed, edge, spec), seed, heuristicSeats(2), turnBudget, wall, true)
				race, push := 0, 0
				for _, m := range res.moves {
					if strings.Contains(m, "two-turn race") {
						race++
					}
					if strings.Contains(m, "all-in for the kill") {
						push++
					}
				}
				t.Logf("edge seat %d: state=%s turns=%d winner=%d lives=%v in %v; %d race and %d all-in attacks",
					edge, res.state, res.turns, res.winner, res.lives, res.elapsed, race, push)
				if res.state != game.StateEnded {
					t.Fatalf("the Wurm edge was not cashed inside %d turns (state %s, lives %v); the board it stopped on:%s",
						turnBudget, res.state, res.lives, res.board)
				}
				if res.winner != edge {
					t.Errorf("seat %d won; the seat with the extra Wurm is seat %d (lives %v)", res.winner, edge, res.lives)
				}
				assertNoEnumeratorBugs(t, res)
			})
		}
	}
}

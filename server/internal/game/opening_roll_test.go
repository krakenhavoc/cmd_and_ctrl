package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// opening_roll_test.go — ADR 0121 §1–§2 and §9: the opening roll as a
// pre-game window.

// openOpeningRollForTest seats n players with distinct decks and opens
// the opening roll on the seeded source, as startAutomaticForKAT starts
// the automatic one.
func openOpeningRollForTest(t *testing.T, n int, seed1, seed2 uint64) *Game {
	t.Helper()
	g := NewGame()
	for seat := 0; seat < n; seat++ {
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", seat+1), buildTestDeck(fmt.Sprintf("Commander %d", seat+1))); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithOpeningRoll(rand.New(rand.NewPCG(seed1, seed2))); err != nil {
		t.Fatalf("StartWithOpeningRoll: %v", err)
	}
	return g
}

// rollUntilChooser rolls every owed die until a chooser exists, taking
// the seats of each round in the order `order` gives (a permutation of
// the round's seats), so the test can roll in an order the automatic
// path never uses.
func rollUntilChooser(t *testing.T, g *Game, order func(round int, seats []int) []int) {
	t.Helper()
	for i := 0; g.OpeningRoll.Chooser < 0; i++ {
		if i > 64 {
			t.Fatal("the opening roll did not settle in 64 rounds")
		}
		round := g.OpeningRoll.Rounds[len(g.OpeningRoll.Rounds)-1]
		for _, seat := range order(len(g.OpeningRoll.Rounds), append([]int(nil), round.Seats...)) {
			if err := g.RollOpening(g.Seats[seat].ID); err != nil {
				t.Fatalf("RollOpening seat %d: %v", seat, err)
			}
		}
	}
}

func reversed(_ int, seats []int) []int {
	for i, j := 0, len(seats)-1; i < j; i, j = i+1, j-1 {
		seats[i], seats[j] = seats[j], seats[i]
	}
	return seats
}

// ADR 0121 §1, "the same winner either way": for the same key, rolling
// by hand in any order finds the chooser the automatic roll finds, and
// the hands and libraries dealt after the choice are the ones the
// automatic path deals — whoever is chosen to go first.
func TestOpeningRollFindsTheAutomaticChooserAndDealsTheSameCards(t *testing.T) {
	for _, n := range []int{2, 3, 4} {
		for s := uint64(1); s <= 60; s++ {
			auto := startAutomaticForKAT(t, n, s, 2026+s)

			byHand := openOpeningRollForTest(t, n, s, 2026+s)
			rollUntilChooser(t, byHand, reversed)
			chooser := byHand.OpeningRoll.Chooser
			if chooser != auto.StartingSeat {
				t.Fatalf("%d seats, seed %d: chooser = seat %d, the automatic roll's winner is seat %d", n, s, chooser, auto.StartingSeat)
			}
			if got, want := openingRollsSorted(byHand), openingRollsSorted(auto); got != want {
				t.Fatalf("%d seats, seed %d: dice by seat = %s, automatic = %s", n, s, got, want)
			}
			// Give the first turn away: the libraries do not care.
			chosen := (chooser + 1) % n
			if err := byHand.ChooseStartingPlayer(byHand.Seats[chooser].ID, chosen); err != nil {
				t.Fatalf("ChooseStartingPlayer: %v", err)
			}
			if byHand.StartingSeat != chosen || byHand.Turn.ActiveSeat != chosen || byHand.Turn.Seq != 1 {
				t.Fatalf("after choosing seat %d: starting seat %d, turn %+v", chosen, byHand.StartingSeat, byHand.Turn)
			}
			if got, want := openingZonesDigest(byHand), openingZonesDigest(auto); got != want {
				t.Fatalf("%d seats, seed %d: hands and libraries differ from the automatic path's", n, s)
			}
		}
	}
}

// openingRollsSorted is every opening die grouped by seat, in the order
// each seat rolled them: "seat 0: 7 20 | seat 1: 4 …". Rolling order
// across seats is the one thing the two paths may disagree on.
func openingRollsSorted(g *Game) string {
	bySeat := make([][]int, len(g.Seats))
	for _, ev := range g.Events {
		if ev.Kind != EventRollDie {
			continue
		}
		for i, p := range g.Seats {
			if p.ID == ev.Actor {
				bySeat[i] = append(bySeat[i], ev.Amount)
			}
		}
	}
	return fmt.Sprint(bySeat)
}

func TestOpeningRollWindowShape(t *testing.T) {
	g := openOpeningRollForTest(t, 4, 22, 2048)
	if g.State != StateActive || !g.MulligansOpen || g.OpeningRoll == nil {
		t.Fatalf("state %s, mulligans open %v, opening roll %+v", g.State, g.MulligansOpen, g.OpeningRoll)
	}
	if g.Turn.Seq != 0 || g.Turn.Step != StepUntap || g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("turn = %+v, want the pre-game turn (Seq 0, untap, no priority)", g.Turn)
	}
	for i, p := range g.Seats {
		if len(p.Hand.Cards) != 0 {
			t.Fatalf("seat %d holds %d cards before the choice", i, len(p.Hand.Cards))
		}
		if p.Life != g.Settings.StartingLife {
			t.Fatalf("seat %d life = %d, want %d", i, p.Life, g.Settings.StartingLife)
		}
	}
	if !reflect.DeepEqual(g.OpeningRoll.Rounds[0].Seats, []int{0, 1, 2, 3}) || g.OpeningRoll.Chooser != -1 {
		t.Fatalf("opening roll = %+v", g.OpeningRoll)
	}
	if err := g.KeepHand(g.Seats[0].ID); !errors.Is(err, ErrOpeningRollOpen) {
		t.Fatalf("KeepHand during the roll: %v, want ErrOpeningRollOpen", err)
	}
	if err := g.Mulligan(g.Seats[0].ID, 7); !errors.Is(err, ErrOpeningRollOpen) {
		t.Fatalf("Mulligan during the roll: %v, want ErrOpeningRollOpen", err)
	}
}

// Key (22, 2048) at four seats: round 1 is 7, 4, 8, 8, so seats 2 and 3
// roll again, and seat 3 wins round 2 with 18 to seat 2's 9.
func TestOpeningRollTieOpensARoundOfTheLeadersOnly(t *testing.T) {
	g := openOpeningRollForTest(t, 4, 22, 2048)
	for _, seat := range []int{3, 1, 0} {
		if err := g.RollOpening(g.Seats[seat].ID); err != nil {
			t.Fatalf("RollOpening seat %d: %v", seat, err)
		}
	}
	if err := g.RollOpening(g.Seats[3].ID); !errors.Is(err, ErrNotYourRoll) {
		t.Fatalf("a second die in one round: %v, want ErrNotYourRoll", err)
	}
	if len(g.OpeningRoll.Rounds) != 1 {
		t.Fatalf("a round ended before every seat rolled: %+v", g.OpeningRoll)
	}
	if err := g.RollOpening(g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	or := g.OpeningRoll
	if len(or.Rounds) != 2 || !reflect.DeepEqual(or.Rounds[1].Seats, []int{2, 3}) || or.Chooser != -1 {
		t.Fatalf("after the tie: %+v, want round 2 of seats 2 and 3", or)
	}
	tie := lastEventOf(g, EventOpeningRoll)
	if tie.Label != OpeningRollTie || !reflect.DeepEqual(tie.Seats, []int{2, 3}) || tie.Amount != 8 {
		t.Fatalf("tie event = %+v", tie)
	}
	for _, seat := range []int{0, 1} {
		if err := g.RollOpening(g.Seats[seat].ID); !errors.Is(err, ErrNotYourRoll) {
			t.Fatalf("seat %d rolled outside the round: %v, want ErrNotYourRoll", seat, err)
		}
	}
	if err := g.ChooseStartingPlayer(g.Seats[2].ID, 2); !errors.Is(err, ErrNoChooserYet) {
		t.Fatalf("a choice before the roll is won: %v, want ErrNoChooserYet", err)
	}
	for _, seat := range []int{2, 3} {
		if err := g.RollOpening(g.Seats[seat].ID); err != nil {
			t.Fatal(err)
		}
	}
	if g.OpeningRoll.Chooser != 3 {
		t.Fatalf("chooser = %d, want seat 3", g.OpeningRoll.Chooser)
	}
	won := lastEventOf(g, EventOpeningRoll)
	if won.Label != OpeningRollWon || won.Actor != g.Seats[3].ID || won.Amount != 18 {
		t.Fatalf("won event = %+v", won)
	}
	if err := g.ChooseStartingPlayer(g.Seats[2].ID, 2); !errors.Is(err, ErrNotChooser) {
		t.Fatalf("the loser chose: %v, want ErrNotChooser", err)
	}
	if err := g.ChooseStartingPlayer(g.Seats[3].ID, 9); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("an unseated choice: %v, want ErrInvalidParam", err)
	}
	if err := g.ChooseStartingPlayer(g.Seats[3].ID, 1); err != nil {
		t.Fatal(err)
	}
	if g.OpeningRoll != nil || g.StartingSeat != 1 || g.Turn.ActiveSeat != 1 || g.Turn.Seq != 1 || !g.MulligansOpen {
		t.Fatalf("after the choice: roll %+v, starting seat %d, turn %+v, mulligans %v", g.OpeningRoll, g.StartingSeat, g.Turn, g.MulligansOpen)
	}
	for i, p := range g.Seats {
		if len(p.Hand.Cards) != OpeningHandSize {
			t.Fatalf("seat %d holds %d cards after the choice", i, len(p.Hand.Cards))
		}
		for _, c := range p.Hand.Cards {
			if !c.IsKnownTo(p.ID) {
				t.Fatalf("seat %d does not know its own card %s", i, c.Name)
			}
		}
	}
	chose := lastEventOf(g, EventStartingPlayer)
	if chose.Actor != g.Seats[3].ID || chose.Target != g.Seats[1].ID {
		t.Fatalf("starting player event = %+v", chose)
	}
	if err := g.RollOpening(g.Seats[0].ID); !errors.Is(err, ErrNoOpeningRoll) {
		t.Fatalf("a roll after the window closed: %v, want ErrNoOpeningRoll", err)
	}
	// CR 103.5: the chosen starting seat (1) decides first, so seat 0
	// waits for its turn.
	if err := g.KeepHand(g.Seats[0].ID); !errors.Is(err, ErrNotYourMulligan) {
		t.Fatalf("KeepHand out of turn: %v, want ErrNotYourMulligan", err)
	}
	if err := g.KeepHand(g.Seats[1].ID); err != nil {
		t.Fatalf("KeepHand after the choice: %v", err)
	}
}

func lastEventOf(g *Game, kind EventKind) Event {
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == kind {
			return g.Events[i]
		}
	}
	return Event{}
}

// §3: the host's button rolls only the CURRENT round's missing seats,
// in seat order, with the presser recorded, and it cannot change a
// result — only when it is shown.
func TestHostRollRemainingRollsTheCurrentRoundOnly(t *testing.T) {
	g := openOpeningRollForTest(t, 4, 22, 2048)
	if err := g.RollOpening(g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	host := g.Seats[0].ID
	if err := g.HostRollRemaining(host); err != nil {
		t.Fatal(err)
	}
	round1 := g.OpeningRoll.Rounds[0]
	want := []OpeningRollDie{{Seat: 2, Result: 8, By: 2}, {Seat: 0, Result: 7, By: 0}, {Seat: 1, Result: 4, By: 0}, {Seat: 3, Result: 8, By: 0}}
	if !reflect.DeepEqual(round1.Rolls, want) {
		t.Fatalf("round 1 = %+v, want %+v", round1.Rolls, want)
	}
	var rolledFor Event
	for _, ev := range g.Events {
		if ev.Kind == EventOpeningRoll && ev.Label == OpeningRollRolledFor {
			rolledFor = ev
		}
	}
	if rolledFor.Actor != host || !reflect.DeepEqual(rolledFor.Seats, []int{1, 3}) {
		t.Fatalf("rolled-for event = %+v, want the host rolling for seats 1 and 3", rolledFor)
	}
	// The tie it caused has its own dice to roll; the admin's button
	// rolls them, recorded as nobody's seat.
	if len(g.OpeningRoll.Rounds) != 2 || len(g.OpeningRoll.Rounds[1].Rolls) != 0 {
		t.Fatalf("the host's button rolled into the next round: %+v", g.OpeningRoll)
	}
	if err := g.HostRollRemaining(uuid.Nil); err != nil {
		t.Fatal(err)
	}
	round2 := g.OpeningRoll.Rounds[1].Rolls
	if len(round2) != 2 || round2[0].By != OpeningRollByAdmin || round2[1].By != OpeningRollByAdmin {
		t.Fatalf("round 2 = %+v, want both rolled by the admin", round2)
	}
	if g.OpeningRoll.Chooser != 3 {
		t.Fatalf("chooser = %d, want seat 3 (the same winner the seats would have rolled)", g.OpeningRoll.Chooser)
	}
	if err := g.HostRollRemaining(host); !errors.Is(err, ErrNobodyLeftToRoll) {
		t.Fatalf("the button with nobody left: %v, want ErrNobodyLeftToRoll", err)
	}
}

// §1: a concession during the roll strikes the seat and reads the
// standings again, the chooser's included; the last seat standing wins.
func TestConcedeDuringTheOpeningRoll(t *testing.T) {
	t.Run("a tied leader leaves and the other chooses", func(t *testing.T) {
		g := openOpeningRollForTest(t, 4, 22, 2048)
		if err := g.HostRollRemaining(uuid.Nil); err != nil { // round 1: 7 4 8 8
			t.Fatal(err)
		}
		if err := g.Concede(g.Seats[2].ID); err != nil {
			t.Fatal(err)
		}
		or := g.OpeningRoll
		if or == nil || or.Chooser != 3 || !reflect.DeepEqual(or.Rounds[1].Seats, []int{3}) {
			t.Fatalf("after seat 2 left: %+v, want seat 3 the chooser", or)
		}
		for _, r := range or.Rounds {
			for _, d := range r.Rolls {
				if d.Seat == 2 {
					t.Fatalf("seat 2's die is still in %+v", r)
				}
			}
		}
		if g.Turn.Seq != 0 || g.Turn.ActiveSeat != 0 || !g.Seats[2].Eliminated {
			t.Fatalf("the departure moved the pre-game turn: %+v", g.Turn)
		}
	})
	t.Run("the chooser leaves and the rest roll again", func(t *testing.T) {
		g := openOpeningRollForTest(t, 4, 22, 2048)
		if err := g.HostRollRemaining(uuid.Nil); err != nil {
			t.Fatal(err)
		}
		if err := g.HostRollRemaining(uuid.Nil); err != nil { // round 2: seat 3 wins
			t.Fatal(err)
		}
		if g.OpeningRoll.Chooser != 3 {
			t.Fatalf("chooser = %d", g.OpeningRoll.Chooser)
		}
		if err := g.Concede(g.Seats[3].ID); err != nil {
			t.Fatal(err)
		}
		// Round 2 is seat 2 alone: it chooses on its round-2 die.
		if g.OpeningRoll.Chooser != 2 {
			t.Fatalf("after the chooser left: %+v, want seat 2 the chooser", g.OpeningRoll)
		}
		won := lastEventOf(g, EventOpeningRoll)
		if won.Label != OpeningRollWon || won.Actor != g.Seats[2].ID {
			t.Fatalf("won event = %+v", won)
		}
	})
	t.Run("a winner by round 1 leaves", func(t *testing.T) {
		g := openOpeningRollForTest(t, 4, 2, 2028) // 1 6 5 15: seat 3 wins
		if err := g.HostRollRemaining(uuid.Nil); err != nil {
			t.Fatal(err)
		}
		if g.OpeningRoll.Chooser != 3 {
			t.Fatalf("chooser = %d, want 3", g.OpeningRoll.Chooser)
		}
		if err := g.Concede(g.Seats[3].ID); err != nil {
			t.Fatal(err)
		}
		if g.OpeningRoll.Chooser != 1 { // 6 is the high among the rest
			t.Fatalf("after the chooser left: %+v, want seat 1", g.OpeningRoll)
		}
	})
	t.Run("the last seat standing wins", func(t *testing.T) {
		g := openOpeningRollForTest(t, 2, 1, 2027)
		if err := g.Concede(g.Seats[0].ID); err != nil {
			t.Fatal(err)
		}
		if g.State != StateEnded || g.OpeningRoll != nil || g.Outcome == nil || g.Outcome.Winner != g.Seats[1].ID {
			t.Fatalf("state %s, roll %+v, outcome %+v", g.State, g.OpeningRoll, g.Outcome)
		}
	})
	t.Run("a seat that left cannot roll or be chosen", func(t *testing.T) {
		g := openOpeningRollForTest(t, 3, 1, 2027) // 20 17 1: seat 0 wins
		if err := g.Concede(g.Seats[2].ID); err != nil {
			t.Fatal(err)
		}
		if err := g.RollOpening(g.Seats[2].ID); !errors.Is(err, ErrPlayerEliminated) {
			t.Fatalf("a departed seat rolled: %v", err)
		}
		if err := g.HostRollRemaining(uuid.Nil); err != nil {
			t.Fatal(err)
		}
		if err := g.ChooseStartingPlayer(g.Seats[0].ID, 2); !errors.Is(err, ErrPlayerEliminated) {
			t.Fatalf("a departed seat was chosen: %v", err)
		}
		if err := g.ChooseStartingPlayer(g.Seats[0].ID, 0); err != nil {
			t.Fatal(err)
		}
		if n := len(g.Seats[2].Hand.Cards); n != 0 {
			t.Fatalf("the departed seat was dealt %d cards", n)
		}
	})
}

// §1, persistence: a game captured mid-roll restores mid-roll, and its
// next die is the one the live game rolls next.
func TestOpeningRollSurvivesASnapshotAndAClone(t *testing.T) {
	live := openOpeningRollForTest(t, 4, 22, 2048)
	if err := live.HostRollRemaining(live.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(live.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if !reflect.DeepEqual(restored.OpeningRoll, live.OpeningRoll) {
		t.Fatalf("restored roll = %+v, live = %+v", restored.OpeningRoll, live.OpeningRoll)
	}

	undo := live.Clone()
	for _, g := range []*Game{live, restored} {
		if err := g.RollOpening(g.Seats[2].ID); err != nil {
			t.Fatal(err)
		}
	}
	if a, b := live.OpeningRoll.Rounds[1].Rolls, restored.OpeningRoll.Rounds[1].Rolls; !reflect.DeepEqual(a, b) {
		t.Fatalf("the next die differs after a restore: live %+v, restored %+v", a, b)
	}
	if len(undo.OpeningRoll.Rounds[1].Rolls) != 0 {
		t.Fatalf("the clone shares the live rounds: %+v", undo.OpeningRoll)
	}
	live.WithWriteLock(func() { live.RestoreFrom(undo) })
	if len(live.OpeningRoll.Rounds[1].Rolls) != 0 {
		t.Fatalf("RestoreFrom did not put the roll back: %+v", live.OpeningRoll)
	}
}

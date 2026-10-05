package aiseat_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// first_turn_test.go — ADR 0125 §5.2: the practice table rolls for the
// first turn, and a practice bot that wins the roll hands the first
// turn to the player through the ordinary choose_starting_player move.
// A bot with no FirstTurnTo chooses as ADR 0121 §4 has it.

func chooseMove(seat int, self bool) legal.Move {
	params, _ := json.Marshal(map[string]int{"seat": seat})
	label := fmt.Sprintf("Seat %d goes first", seat)
	if self {
		label = "I go first"
	}
	return legal.Move{Type: legal.TypeChooseStartingPlayer, Kind: legal.KindOpeningRoll, Label: label, Params: params, AlwaysLegal: self}
}

func TestFirstTurnIndex(t *testing.T) {
	roll := legal.Move{Type: legal.TypeRollOpening, Kind: legal.KindOpeningRoll, AlwaysLegal: true}
	pass := legal.Move{Type: legal.TypePassPriority, Kind: legal.KindPass, AlwaysLegal: true}
	for _, tc := range []struct {
		name  string
		moves []legal.Move
		seat  int
		want  int
	}{
		{"the winner's choice names the seat", []legal.Move{chooseMove(0, false), chooseMove(1, true)}, 0, 0},
		{"the seat in another position", []legal.Move{chooseMove(1, true), chooseMove(0, false)}, 0, 1},
		{"a die to roll is the policy's", []legal.Move{roll}, 0, -1},
		{"a mixed window is not the choice", []legal.Move{chooseMove(0, false), pass}, 0, -1},
		{"a seat not offered", []legal.Move{chooseMove(1, true), chooseMove(2, false)}, 0, -1},
		{"no moves", nil, 0, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := aiseat.FirstTurnIndex(tc.moves, tc.seat); got != tc.want {
				t.Errorf("FirstTurnIndex = %d, want %d", got, tc.want)
			}
		})
	}
}

// firstTurnTally records, from the runners' own reports, every
// opening-roll move a bot applied (#634: a complete history, so every
// count is monotone and safe to wait on).
type firstTurnTally struct {
	mu       sync.Mutex
	rolls    int
	choices  []aiseat.DecisionEvent
	rejected []string
}

func (o *firstTurnTally) Observe(ev aiseat.DecisionEvent) {
	if ev.Index < 0 || ev.Index >= len(ev.Input.Moves) {
		return
	}
	mv := ev.Input.Moves[ev.Index]
	if mv.Kind != legal.KindOpeningRoll {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if !ev.Applied {
		o.rejected = append(o.rejected, fmt.Sprintf("%s: %v", mv.Label, ev.RejectErr))
		return
	}
	switch mv.Type {
	case legal.TypeRollOpening:
		o.rolls++
	case legal.TypeChooseStartingPlayer:
		o.choices = append(o.choices, ev)
	}
}

func (o *firstTurnTally) snapshot() (int, []aiseat.DecisionEvent, []string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.rolls, append([]aiseat.DecisionEvent(nil), o.choices...), append([]string(nil), o.rejected...)
}

// twoSeatRollGame is a practice-shaped table: the player at seat 0 and
// a bot at seat 1, started with the opening roll open.
func twoSeatRollGame(t *testing.T, s1, s2 uint64) *game.Game {
	t.Helper()
	g := game.NewGame()
	for _, name := range []string{"Alice", "Practice Bot"} {
		if _, err := g.AddPlayer(name, monoRedDeck(uuid.Nil)); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithOpeningRoll(rand.New(rand.NewPCG(s1, s2))); err != nil {
		t.Fatalf("StartWithOpeningRoll: %v", err)
	}
	return g
}

// seedWonBy finds a key whose opening roll seat wins. A seat's k-th
// opening die depends only on the key, the seat and k (ADR 0121 §1,
// "the same winner either way"), so the table built from the same key
// finds the same winner whoever rolls first, a bot on its own clock
// included.
func seedWonBy(t *testing.T, seat int) uint64 {
	t.Helper()
	for s := uint64(1); s < 200; s++ {
		g := twoSeatRollGame(t, s, s+1)
		for g.OpeningRoll.Chooser < 0 {
			round := g.OpeningRoll.Rounds[len(g.OpeningRoll.Rounds)-1]
			for _, rs := range round.Seats {
				if err := g.RollOpening(g.Seats[rs].ID); err != nil {
					t.Fatalf("RollOpening: %v", err)
				}
			}
		}
		if g.OpeningRoll.Chooser == seat {
			return s
		}
	}
	t.Fatalf("no key in 200 makes seat %d win the opening roll", seat)
	return 0
}

// actAsPlayer makes the player's own opening-roll move, if they have
// one: their die, or, as the winner, the first turn for themselves. It
// goes through the room, as the hub's dispatch does, so the bot's
// runner is woken by it.
func actAsPlayer(t *testing.T, room *ws.Room, player uuid.UUID) {
	t.Helper()
	g := room.Game
	for _, mv := range legal.EnumerateFor(g, player) {
		switch {
		case mv.Type == legal.TypeRollOpening:
			if _, _, err := room.ApplyExternal(func() error { return g.RollOpening(player) }); err != nil {
				t.Fatalf("the player's roll: %v", err)
			}
			return
		case mv.Type == legal.TypeChooseStartingPlayer && mv.AlwaysLegal:
			seat := g.PlayerByID(player).Seat
			if _, _, err := room.ApplyExternal(func() error { return g.ChooseStartingPlayer(player, seat) }); err != nil {
				t.Fatalf("the player's choice: %v", err)
			}
			return
		}
	}
}

// playOpeningRoll plays the player's half of the opening roll until the
// window closes, which is monotone: it never reopens.
func playOpeningRoll(t *testing.T, room *ws.Room, player uuid.UUID) {
	t.Helper()
	waitFor(t, "the opening roll to close", func() bool {
		if !room.Game.OpeningRollOpen() {
			return true
		}
		actAsPlayer(t, room, player)
		return !room.Game.OpeningRollOpen()
	})
}

func startPracticeBot(t *testing.T, room *ws.Room, firstTurnTo *int) (*aiseat.Manager, *firstTurnTally) {
	t.Helper()
	tally := &firstTurnTally{}
	host := aiseat.NewManagerWithConfig(nil, aiseat.Config{Observer: tally}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	host.StartBots(room, []aiseat.SeatSpec{{
		PlayerID:    room.Game.Seats[1].ID,
		Tier:        string(aiseat.TierRandom),
		FirstTurnTo: firstTurnTo,
	}})
	t.Cleanup(func() { host.StopBots(room.Game.ID) })
	return host, tally
}

func TestPracticeBotHandsTheFirstTurnToThePlayer(t *testing.T) {
	player := 0

	t.Run("the bot wins", func(t *testing.T) {
		s := seedWonBy(t, 1)
		room := ws.NewRoom(twoSeatRollGame(t, s, s+1), testLogger(), "")
		g := room.Game
		_, tally := startPracticeBot(t, room, &player)
		playOpeningRoll(t, room, g.Seats[0].ID)

		var starting int
		g.ReadSnapshot(func() { starting = g.StartingSeat })
		if starting != player {
			t.Errorf("starting seat %d, want the player's, %d", starting, player)
		}
		rolls, choices, rejected := tally.snapshot()
		if len(rejected) != 0 {
			t.Errorf("opening-roll moves rejected: %v", rejected)
		}
		if rolls < 1 {
			t.Errorf("the bot rolled %d dice, want its own", rolls)
		}
		// The choice went through the runner's ordinary path: an
		// offered choose_starting_player, applied, and reported to the
		// observer like any other window.
		if len(choices) != 1 {
			t.Fatalf("the bot made %d choices, want 1: %+v", len(choices), choices)
		}
		ev := choices[0]
		mv := ev.Input.Moves[ev.Index]
		if mv.Label != "Alice goes first" || ev.Trace.Rule != aiseat.RuleFirstTurnTo || ev.Trace.Layer != "A" {
			t.Errorf("the bot chose %q (layer %q, rule %q), want Alice by %q", mv.Label, ev.Trace.Layer, ev.Trace.Rule, aiseat.RuleFirstTurnTo)
		}
	})

	t.Run("the player wins", func(t *testing.T) {
		s := seedWonBy(t, 0)
		room := ws.NewRoom(twoSeatRollGame(t, s, s+1), testLogger(), "")
		g := room.Game
		_, tally := startPracticeBot(t, room, &player)
		// Monotone: once the player is the chooser, only their choice
		// closes the window.
		waitFor(t, "the player to win the roll", func() bool {
			chooser := -1
			g.ReadSnapshot(func() {
				if g.OpeningRoll != nil {
					chooser = g.OpeningRoll.Chooser
				}
			})
			if chooser < 0 {
				actAsPlayer(t, room, g.Seats[0].ID)
			}
			return chooser == 0
		})
		if moves := legal.EnumerateFor(g, g.Seats[1].ID); len(moves) != 0 {
			t.Errorf("the bot is offered %+v while the player chooses", moves)
		}
		playOpeningRoll(t, room, g.Seats[0].ID)
		var starting int
		g.ReadSnapshot(func() { starting = g.StartingSeat })
		if starting != player {
			t.Errorf("starting seat %d, want the player's, %d", starting, player)
		}
		_, choices, rejected := tally.snapshot()
		if len(rejected) != 0 {
			t.Errorf("opening-roll moves rejected: %v", rejected)
		}
		if len(choices) != 0 {
			t.Errorf("the bot chose who goes first at a roll the player won: %+v", choices)
		}
	})
}

// With FirstTurnTo unset, the window is the tier's: a random bot that
// wins answers it through its own policy (ADR 0121 §4), so any table
// but the practice table is unchanged.
func TestFirstTurnToUnsetLeavesTheChoiceToTheTier(t *testing.T) {
	s := seedWonBy(t, 1)
	room := ws.NewRoom(twoSeatRollGame(t, s, s+1), testLogger(), "")
	g := room.Game
	_, tally := startPracticeBot(t, room, nil)
	playOpeningRoll(t, room, g.Seats[0].ID)
	_, choices, _ := tally.snapshot()
	if len(choices) != 1 {
		t.Fatalf("the bot made %d choices, want 1", len(choices))
	}
	if tr := choices[0].Trace; tr.Rule == aiseat.RuleFirstTurnTo || tr.Layer != aiseat.TraceLayerRandom {
		t.Errorf("the choice was answered by layer %q rule %q, want the random tier's own", tr.Layer, tr.Rule)
	}
}

// The whole path through the lobby (ADR 0125 §5.2): a practice table
// opens the opening roll with nothing dealt, and whoever wins, the
// player takes turn one. The key is random (the practice route seeds
// nothing), so tables are opened until both outcomes have been seen.
func TestPracticeTableRollsAndThePlayerTakesTurnOne(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	l := lobby.NewLobby(ws.NewRoomManager(log, ""))
	host := aiseat.NewManagerWithConfig(nil, aiseat.Config{}, log)
	l.SetBotHost(host)
	t.Cleanup(host.Shutdown)

	won := map[int]bool{}
	for i := 0; i < 64 && len(won) < 2; i++ {
		meta, playerID, err := l.CreatePractice(
			lobby.PracticeHuman{Name: "Alice", Owner: "user:alice", DeckName: "First Steps", Cards: monoRedDeck(uuid.Nil)},
			lobby.PracticeBot{Name: "Practice Bot", Tier: string(aiseat.TierRandom), DeckName: "Practice Partner", Cards: monoRedDeck(uuid.Nil)},
		)
		if err != nil {
			t.Fatalf("CreatePractice: %v", err)
		}
		room := l.RoomOf(meta.ID)
		g := room.Game
		if i == 0 {
			var open bool
			var seq, hand int
			g.ReadSnapshot(func() {
				open = g.OpeningRoll != nil
				seq = g.Turn.Seq
				for _, p := range g.Seats {
					hand += len(p.Hand.Cards)
				}
			})
			if !open || seq != 0 || hand != 0 {
				t.Fatalf("a new practice table: opening roll open %v, turn %d, %d cards in hands; want the roll open and nothing dealt", open, seq, hand)
			}
		}
		playOpeningRoll(t, room, playerID)
		// The winner is in the log: the bot may win and choose between
		// two polls, so the window's own Chooser can be gone already.
		chooser, starting := -1, -1
		g.ReadSnapshot(func() {
			starting = g.StartingSeat
			for _, ev := range g.Events {
				if ev.Kind == game.EventOpeningRoll && ev.Label == game.OpeningRollWon {
					for _, p := range g.Seats {
						if p.ID == ev.Actor {
							chooser = p.Seat
						}
					}
				}
			}
		})
		if chooser < 0 {
			t.Fatalf("table %d: no winner in the log", i)
		}
		if starting != 0 {
			t.Fatalf("table %d: seat %d won the roll and seat %d takes turn one, want the player (seat 0)", i, chooser, starting)
		}
		won[chooser] = true
	}
	if !won[0] || !won[1] {
		t.Fatalf("winners seen: %v; want both the player and the bot", won)
	}
}

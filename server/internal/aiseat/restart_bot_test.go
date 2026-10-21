package aiseat_test

import (
	"context"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// restart_bot_test.go — #3856, CR 727: a table of bots plays through a
// restarted game. The game restarts under the runners' feet (as Karn
// Liberated's −14 would, mid-game, with a card left in exile); every
// bot seat then answers the new game's mulligan window and plays on,
// the restarter takes the first turn, and the exempted card enters
// under the restarter's control. Not behind AISEAT_GAME_TESTS: it
// plays a few turns, not a game.
func TestBotsPlayOnThroughARestartedGame(t *testing.T) {
	room := newRoom(t, 4, 3856)
	g := room.Game
	meter := &rules.Meter{}
	policies := []aiseat.Policy{
		rules.NewFilter(heuristic.New(), meter),
		aiseat.NewRandomPolicy(rand.NewPCG(38, 56)),
		rules.NewFilter(heuristic.New(), meter),
		heuristic.New(),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runners := make([]*aiseat.Runner, 0, len(policies))
	for i, p := range g.Seats {
		runners = append(runners, startBot(t, ctx, room, p.ID, policies[i], aiseat.Config{}, nil, testLogger()))
	}

	waitFor(t, "the bots to reach the old game's second turn", func() bool {
		done := false
		g.ReadSnapshot(func() {
			done = g.State != game.StateActive || (!g.MulligansOpen && g.Turn.Seq >= 2)
		})
		return done
	})

	restarter, owner := g.Seats[2].ID, g.Seats[1].ID
	exempt := game.NewCard("Karn's Prize", owner)
	exempt.TypeLine = "Creature — Bear"
	exempt.Power, exempt.Toughness = 2, 2
	var before uint64
	restarted := false
	if _, _, err := room.ApplyExternal(func() error {
		before = g.LastEventSeq()
		g.WithWriteLock(func() {
			if g.State != game.StateActive {
				return
			}
			restarted = true
			g.Exile.PushTop(exempt)
			g.RestartTheGameForEffect(restarter, uuid.Nil, []uuid.UUID{exempt.InstanceID})
		})
		g.SettleResolution()
		return nil
	}); err != nil {
		t.Fatalf("restart: %v", err)
	}
	if !restarted {
		t.Skip("the old game ended before the restart could happen")
	}

	// Monotone from here: the new game's mulligan window only closes
	// and its turns only begin.
	waitFor(t, "the bots to mulligan into the new game and reach its third turn", func() bool {
		done := false
		g.ReadSnapshot(func() {
			done = g.State != game.StateActive || (!g.MulligansOpen && g.Turn.Seq >= 3)
		})
		return done
	})
	cancel()
	for _, r := range runners {
		waitForRunner(t, "the runner to exit", r)
	}

	g.ReadSnapshot(func() {
		if g.StartingSeat != 2 {
			t.Errorf("the new game's starting seat is %d, want the restarter's (CR 727.1a)", g.StartingSeat)
		}
		restarts := 0
		entered := false
		for _, ev := range g.Events {
			if ev.Seq <= before {
				continue
			}
			if ev.Kind == game.EventGameRestarted {
				restarts++
			}
			if ev.Kind == game.EventETB {
				if c, ok := g.LookupCardForEffect(ev.CardID); ok && c.Name == "Karn's Prize" {
					entered = true
				}
			}
		}
		if restarts != 1 {
			t.Errorf("%d restart events, want 1", restarts)
		}
		if !entered {
			t.Error("the exempted card never entered the battlefield (CR 727.4)")
		}
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Karn's Prize" && c.Controller != restarter {
				t.Errorf("the exempted card is controlled by %s, want the restarter", c.Controller)
			}
		}
	})
}

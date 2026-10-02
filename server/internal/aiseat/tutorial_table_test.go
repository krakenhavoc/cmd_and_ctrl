package aiseat_test

import (
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// TestTutorialTablePlaysSafely plays the tutorial's practice table
// (ADR 0076 §2.2, #1078) as the lobby builds it — the two fixed
// tutorial decks, the player at seat 0 taking turn one — through the
// first tutorialRounds rounds, and holds it to what the ADR asks of it:
//
//   - nothing stalls. §3: "if aiseat stalls, the tutorial stalls at
//     step 9 in front of exactly the audience least equipped to
//     understand why". The decks are fixed and the seeds are pinned, so
//     unlike the catalog soak a stall here is a regression on a stable
//     input and fails the test.
//   - no card's effect errors.
//   - the player is nowhere near dying. "A practice bot that kills the
//     player during the tutorial is a bug" — the decklist is the fix,
//     and this is its check. The tutorial is about three turns; the
//     window is wider.
//
// Seat 0 stands in for the human with the random policy too, as
// #1073's soak did: a seat that only ever passes never makes the
// interleaving a real table has.
//
// Gated like the catalog soak: AISEAT_GAME_TESTS, and
// CMDCTRL_SCRYFALL_DUMP for the cards. AISEAT_TUTORIAL_GAMES widens the
// sample (default 10), AISEAT_TUTORIAL_SEED moves it (default 7600).
func TestTutorialTablePlaysSafely(t *testing.T) {
	requireGameTests(t)
	dump := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if dump == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to load the tutorial decks")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(dump); err != nil {
		t.Fatalf("load scryfall dump %s: %v", dump, err)
	}
	playerList, err := decks.TutorialPlayer().Load(idx)
	if err != nil {
		t.Fatalf("player deck: %v", err)
	}
	botList, err := decks.TutorialBot().Load(idx)
	if err != nil {
		t.Fatalf("bot deck: %v", err)
	}

	games := 10
	if v, err := strconv.Atoi(os.Getenv("AISEAT_TUTORIAL_GAMES")); err == nil && v > 0 {
		games = v
	}
	base := uint64(7600)
	if v, err := strconv.ParseUint(os.Getenv("AISEAT_TUTORIAL_SEED"), 10, 64); err == nil {
		base = v
	}
	const tutorialRounds = 6
	// Forty life, a bot whose biggest body is a 3/3: anything under
	// this by round six means the deck is not as slow as it claims.
	const lifeFloor = 30
	wall := envDuration("AISEAT_WALLCLOCK", 120*time.Second)
	stats := map[string]*cardStat{}

	for i := 0; i < games; i++ {
		seed := base + uint64(i)
		t.Run(fmt.Sprintf("seed=%d", seed), func(t *testing.T) {
			g := game.NewGame()
			human, err := g.AddPlayer("Player", playerList.ToGameCards())
			if err != nil {
				t.Fatalf("AddPlayer: %v", err)
			}
			if _, err := g.AddPlayer("Practice Bot", botList.ToGameCards()); err != nil {
				t.Fatalf("AddPlayer bot: %v", err)
			}
			// Start, not StartWithFirstPlayerRoll: the lobby seats the
			// player first (lobby.CreatePractice).
			if err := g.Start(rand.New(rand.NewPCG(seed, seed+1))); err != nil {
				t.Fatalf("Start: %v", err)
			}
			room := ws.NewRoom(g, testLogger(), "")
			policies := []aiseat.Policy{
				aiseat.NewRandomPolicy(rand.NewPCG(seed, 0)),
				aiseat.NewRandomPolicy(rand.NewPCG(seed, 1)),
			}
			res := playCatalogGame(t, room, policies, tutorialRounds, wall, nil)
			errs := tally(room.Game, stats)

			if res.stalled {
				t.Errorf("the practice table stalled: %s", res.dump)
			}
			for _, e := range errs {
				t.Errorf("effect error: %s: %s", e.Card, e.Msg)
			}
			var life int
			var out bool
			room.Game.ReadSnapshot(func() { life, out = human.Life, human.Eliminated })
			if out || life < lifeFloor {
				t.Errorf("by round %d the player is at %d life (eliminated=%v); the practice bot is meant to be harmless",
					res.turns, life, out)
			}
			t.Logf("seed %d: state=%s rounds=%d applied=%d in %v", seed, res.state, res.turns, res.applied, res.elapsed)
		})
	}
}

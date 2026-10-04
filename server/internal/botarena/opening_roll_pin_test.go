package botarena_test

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
)

// opening_roll_pin_test.go — ADR 0121 §1 and §9: the arena keeps the
// automatic opening roll (StartWithFirstPlayerRoll), which now runs on
// the same window a person rolls in, with the shuffle and the deal moved
// after the choice. A seeded arena game must be the same game it was
// before that change: the same starting seat, the same opening hands
// and therefore the same moves, turn for turn, to the same result.
//
// The digests below were captured from the engine BEFORE ADR 0121
// (origin/develop at dc4f1844) and must never be regenerated. A
// lockstep game is a pure function of its seed (#1503), so its move log
// is the whole game: a different starting seat, a different hand or a
// different library order changes it.
func TestArenaSeededGameIsTheSameGameAfterTheOpeningRollWindow(t *testing.T) {
	cases := []struct {
		name   string
		cfg    botarena.Config
		moves  int
		turns  int
		winner int
		digest string
	}{
		{
			name: "four seats, two rounds",
			cfg: botarena.Config{
				Seats: []botarena.SeatSpec{
					{Tier: tiers.Heuristic}, {Tier: tiers.Random},
					{Tier: tiers.Heuristic}, {Tier: tiers.Random},
				},
				Games: 1, Seed: 1503, TurnBudget: 2, Wall: 2 * time.Minute, Lockstep: true,
			},
			moves: 356, turns: 3, winner: -1, digest: "2478ceb8cbea76b596f61a3456004c30b9553fcd9f042ecbb8fbdb72112397be",
		},
		{
			name: "two seats, eight rounds",
			cfg: botarena.Config{
				Seats: []botarena.SeatSpec{{Tier: tiers.Heuristic}, {Tier: tiers.Heuristic}},
				Games: 1, Seed: 31, TurnBudget: 8, Wall: 2 * time.Minute, Lockstep: true,
			},
			moves: 411, turns: 9, winner: -1, digest: "eb52ed98509e0dfeb2455bf2731114b98d9d119208188b3ce267b4c59005bf30",
		},
		{
			name: "four heuristic seats to a winner",
			cfg: botarena.Config{
				Seats: []botarena.SeatSpec{
					{Tier: tiers.Heuristic}, {Tier: tiers.Heuristic},
					{Tier: tiers.Heuristic}, {Tier: tiers.Heuristic},
				},
				Games: 1, Seed: 107, TurnBudget: 60, Wall: 5 * time.Minute, Lockstep: true,
			},
			moves: 1971, turns: 15, winner: 3, digest: "45805e9a5f0e48125e8c6330c38df972d884df355fef683eac69c852b57acf08",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, lines := playLogged(t, c.cfg)
			if res.Stalled {
				t.Fatalf("the game stalled:\n%s", res.StallDump)
			}
			sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
			got := hex.EncodeToString(sum[:])
			if c.digest == "" {
				t.Fatalf("capture: moves: %d, turns: %d, winner: %d, digest: %q", len(lines), res.Turns, res.Winner, got)
			}
			if len(lines) != c.moves || res.Turns != c.turns || res.Winner != c.winner || got != c.digest {
				t.Fatalf("seed %d is no longer the same game:\n got  %d moves, turn %d, winner %d, digest %s\n want %d moves, turn %d, winner %d, digest %s",
					c.cfg.Seed, len(lines), res.Turns, res.Winner, got, c.moves, c.turns, c.winner, c.digest)
			}
		})
	}
}

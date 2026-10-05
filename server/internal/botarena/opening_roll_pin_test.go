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
//
// One deliberate exception, re-pinned by hand rather than regenerated
// (#2275, 2026-10-05): CR 117.4's passes in succession. In seed 1503 a
// random seat taps a Mountain for mana while holding priority in
// another player's upkeep, after three seats have passed. A mana
// ability is an action (CR 117.3c), so those three seats pass again
// before the step ends — three more "Pass priority" lines and nothing
// else: the same opening, the same hands, every other move identical.
// The other two seeds never act between passes and are unchanged.
//
// A second deliberate exception, re-pinned by hand (#2237, 2026-10-05):
// CR 103.5 mulligans in turn order. Keep and mulligan are now answered
// starting with the starting seat, so the SAME decisions appear in a
// different order, and the log's seat labels (the order of first
// appearance) are renumbered because the starting seat now acts first.
// Checked against the pre-change engine on seed 1503: the same starting
// seat, the same keeps and mulligans (a random seat mulligans to seven,
// then six, then keeps; the heuristic seats keep on four and three
// lands), the same 359 moves and the same turn count. Only the order of
// the opening lines and the labels differ. The digests below are of the
// new order.
//
// It is also a function of the policies, so a change to how the
// heuristic attacks or blocks changes a long game after its opening.
// The four-heuristic case is therefore pinned only up to its first
// declared attack (`opening`): the opening roll, the deal, the
// mulligans and every move of the turns before combat, which is what
// ADR 0121 promises to keep. Its digest is of those 440 lines of the
// pre-ADR-0121 game: #1548 derived it from develop at d9291462, whose
// whole move log still matched the full-game digest kept below (and
// checked it again at ca8e4d29, after #2275's passes in succession), and
// the heuristic's gang blocks and attrition horizon part from that
// game only at move 770, in turn 8's combat.
func TestArenaSeededGameIsTheSameGameAfterTheOpeningRollWindow(t *testing.T) {
	cases := []struct {
		name   string
		cfg    botarena.Config
		moves  int
		turns  int
		winner int
		digest string
		// opening pins only the move log before the first declared
		// attack; moves is its length, and turns and winner are not
		// asked.
		opening bool
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
			moves: 359, turns: 3, winner: -1, digest: "296da15d6f1c0f8953b18f7d41ecedecae065bbcf88b2d0858745a8f7d75dd27",
		},
		{
			name: "two seats, eight rounds",
			cfg: botarena.Config{
				Seats: []botarena.SeatSpec{{Tier: tiers.Heuristic}, {Tier: tiers.Heuristic}},
				Games: 1, Seed: 31, TurnBudget: 8, Wall: 2 * time.Minute, Lockstep: true,
			},
			moves: 411, turns: 9, winner: -1, digest: "cf06761b9b6c7d2c3c67ad4772ac7cc612ced4c30ad60473bfd00a545b128056",
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
			// The whole pre-ADR-0121 game was 1971 moves, turn 15,
			// winner 3, digest 45805e9a5f0e48125e8c6330c38df972d884df355fef683eac69c852b57acf08.
			opening: true, moves: 440, digest: "055cb3a9582a4431b6493a4a54c9bb81d222aa16d4acbc1bf8095bb5022a5a6c",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, lines := playLogged(t, c.cfg)
			if res.Stalled {
				t.Fatalf("the game stalled:\n%s", res.StallDump)
			}
			if c.opening {
				for i, l := range lines {
					if strings.Contains(l, `applied=true "Attack `) {
						lines = lines[:i]
						break
					}
				}
				res.Turns, res.Winner = c.turns, c.winner
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

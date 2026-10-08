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
// A third deliberate exception, re-pinned by hand (#2462, 2026-10-07):
// no attack after the active seat's pass. In seed 31, turn 8, seat 0
// declares its Drake, keeps its Bear home and passes; the lockstep
// schedule then steps seat 0 again while seat 1 holds priority, and the
// enumerator used to offer it "Attack heuristic0 with Bear" as its only
// legal move, which it took (move 270). Now it is offered nothing, the
// Bear stays home, and from move 270 on the game differs: no block, and
// 412 moves to the same turn 9 instead of 411. Checked by diffing the
// two move logs: the first 269 lines are identical.
//
// A fourth deliberate exception, re-pinned by hand (#2690, 2026-10-08):
// a commander that dies in combat costs its next cast's tax. In seed 31,
// turn 8, seat 1's 3/3 commander attacked into seat 0's untapped 3/3
// commander at "blocked at a loss (focus) (+1.54)", and the two traded
// (move 293). With the tax priced the attack is −0.46 and gets no focus
// bonus (#2675), so seat 1 passes, and from move 293 on the game
// differs: 398 moves to the same turn 9 instead of 412. Checked by
// diffing the two move logs, the second with the three knobs of
// #2675, #2690 and #2676 turned off, which still gives the old digest:
// the first 292 lines are identical.
//
// A fifth deliberate exception, re-pinned by hand (#2689, 2026-10-08):
// declared damage at a player is priced per point (ADR 0126's amendment
// of 2026-10-08, D1, `DamageByLethality`). The battle deck's Lightning
// Bolt at a player on 40 life was "cast spell (+1.20)", the flat attack
// price; it is now 3 × DamageToOpponent through the opposition weights,
// less the card, +0.15 at two seats, so the heuristic seat holds it. In
// every seed the first difference is that cast: seed 1503 at move 57
// (346 moves to the same turn 3 instead of 359), seed 31 at move 9 (399
// moves to the same turn 9 instead of 398), and seed 107 at move 14,
// whose first declared attack now comes after 576 moves instead of 440.
// Checked by diffing the move logs with DamageByLethality off, which
// still gives the old digests: every line before those is identical.
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
			moves: 346, turns: 3, winner: -1, digest: "82e0b2b6f5074e68ecb39c9987859470e52da06da4c71c3746e18b8566c71545",
		},
		{
			name: "two seats, eight rounds",
			cfg: botarena.Config{
				Seats: []botarena.SeatSpec{{Tier: tiers.Heuristic}, {Tier: tiers.Heuristic}},
				Games: 1, Seed: 31, TurnBudget: 8, Wall: 2 * time.Minute, Lockstep: true,
			},
			moves: 399, turns: 9, winner: -1, digest: "8571487c58f438604d8572d31045426f4a45eed52c1ce8f86e2fb94b322012ea",
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
			opening: true, moves: 576, digest: "9c1082665a302894a239c9358931e4226d83418bc4d1fc7d5324e004bae96312",
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

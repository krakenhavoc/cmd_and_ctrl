package heuristic_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// combat_pricing_test.go pins the combat slice of the 2026-10-08 review
// games: the focus bonus only on an attack worth making (#2675), an
// attack priced against every block, not only single ones (#2690), and a
// token whose only use is blocking priced by that use (#2676). Each test
// also runs with its knob off, where the old behaviour comes back.

// plant is a 0/1 Plant token, as Tatyova's deck makes them: no power,
// no abilities of any kind.
func plant(id string, controller int, opts ...cardOpt) protocol.CardView {
	c := creature(id, controller, "Plant", 0, 1, opts...)
	c.IsToken = true
	c.TypeLine = "Token Creature — Plant"
	c.ManaCost = ""
	return c
}

// soldier is a 1/1 Soldier token.
func soldier(id string, controller int) protocol.CardView {
	c := creature(id, controller, "Soldier", 1, 1)
	c.IsToken = true
	c.TypeLine = "Token Creature — Soldier"
	c.ManaCost = ""
	return c
}

func withConfig(edit func(*heuristic.Config)) *heuristic.Policy {
	cfg := heuristic.DefaultConfig()
	edit(&cfg)
	return heuristic.NewWithConfig(cfg)
}

// attacksIn reports whether the policy declares an attacker in this
// declare-attackers window, and with what.
func attacksIn(t *testing.T, p *heuristic.Policy, in aiseat.Input) (bool, string) {
	t.Helper()
	d := decide(t, p, in)
	if d.Index < 0 {
		return false, "declined"
	}
	got := chose(t, in, d)
	return got != "Pass priority", got + " — " + d.Reason
}

// #2675, the issue's window: four 0/1 Plants against an untapped 2/4.
// They can deal no damage, so attacking only taps the bot's own
// blockers. FocusBonus used to lift each to +0.55 ("blocked at a loss").
func TestZeroPowerTokensDoNotAttackForTheFocusBonus(t *testing.T) {
	v := newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			plant(cardID(10), 0), plant(cardID(11), 0), plant(cardID(12), 0), plant(cardID(13), 0),
			creature(cardID(20), 1, "Wall", 2, 4),
		),
		withTurn(8, 0, "declare_attackers"),
	)
	moves := []legal.Move{passMove(0)}
	for i := 10; i <= 13; i++ {
		moves = append(moves, attackMove(t, 0, cardID(i), 1))
	}
	in := input(0, v, moves...)
	if ok, got := attacksIn(t, heuristic.New(), in); ok {
		t.Fatalf("attacked: %s; a creature with no power has nothing to attack with", got)
	}
	off := withConfig(func(c *heuristic.Config) { c.FocusNeedsValue = false })
	if ok, got := attacksIn(t, off, in); !ok || !strings.Contains(got, "(focus)") {
		t.Fatalf("with FocusNeedsValue off the Plants attacked as before; got %s", got)
	}
}

// #2675: Plants already declared are not blockers the defender must
// spend. Three of them attacking do not make a Bear "outnumber their
// blockers" when a 2/4 waits to eat it.
func TestDeclaredZeroPowerAttackersDoNotOutnumberTheBlockers(t *testing.T) {
	v := newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			plant(cardID(10), 0, tapped(), attacking(1)),
			plant(cardID(11), 0, tapped(), attacking(1)),
			plant(cardID(12), 0, tapped(), attacking(1)),
			creature(cardID(14), 0, "Bear", 2, 2),
			creature(cardID(20), 1, "Wall", 2, 4),
		),
		withTurn(8, 0, "declare_attackers"),
	)
	in := input(0, v, passMove(0), attackMove(t, 0, cardID(14), 1))
	if ok, got := attacksIn(t, heuristic.New(), in); ok {
		t.Fatalf("attacked: %s; the 2/4 is free to block and kill the Bear", got)
	}
	off := withConfig(func(c *heuristic.Config) { c.FocusNeedsValue = false })
	if ok, got := attacksIn(t, off, in); !ok || !strings.Contains(got, "outnumbers") {
		t.Fatalf("with FocusNeedsValue off the Bear attacked as outnumbering; got %s", got)
	}
}

// #2690, the issue's window (review game 2, seq 317): Mary Read and Anne
// Bonny, a 3/3 commander, into an untapped 2/4 and three 1/1 Soldiers.
// Neither kills her alone, so the single-block check priced the attack at
// +0.54; Y'shtola and a Soldier together kill her and lose the Soldier.
func TestAnAttackerAGangBlockKillsStaysHome(t *testing.T) {
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(34)), newSeat(1, withLife(33))},
		withBattlefield(
			creature(cardID(10), 0, "Mary Read and Anne Bonny", 3, 3, commander()),
			creature(cardID(20), 1, "Y'shtola", 2, 4),
			soldier(cardID(21), 1), soldier(cardID(22), 1), soldier(cardID(23), 1),
		),
		withTurn(7, 0, "declare_attackers"),
	)
	in := input(0, v, passMove(0), attackMove(t, 0, cardID(10), 1))
	if ok, got := attacksIn(t, heuristic.New(), in); ok {
		t.Fatalf("attacked: %s; a gang block kills her for a Soldier", got)
	}
	off := withConfig(func(c *heuristic.Config) { c.GangAwareAttacks = false })
	if ok, got := attacksIn(t, off, in); !ok {
		t.Fatalf("with GangAwareAttacks off she attacked as before; got %s", got)
	}

	// The same attacker into a board whose blockers cannot kill her even
	// together (three 1/1s and nothing else, already short one: two of
	// them free) still attacks.
	open := newView(
		[]protocol.PlayerView{newSeat(0, withLife(34)), newSeat(1, withLife(33))},
		withBattlefield(
			creature(cardID(10), 0, "Mary Read and Anne Bonny", 3, 3, commander()),
			soldier(cardID(21), 1), soldier(cardID(22), 1),
		),
		withTurn(7, 0, "declare_attackers"),
	)
	in = input(0, open, passMove(0), attackMove(t, 0, cardID(10), 1))
	if ok, got := attacksIn(t, heuristic.New(), in); !ok {
		t.Fatalf("two 1/1s cannot kill a 3/3; the attack should go: %s", got)
	}
}

// #2690: a commander that dies costs a recast at {2} more. A 3/3 into
// an even 3/3 trade is worth making with a plain creature (+0.54, then
// the focus bonus) and not with the commander.
func TestACommanderAttackerCountsTheTaxOnItsDeath(t *testing.T) {
	board := func(opts ...cardOpt) protocol.GameView {
		return newView(
			[]protocol.PlayerView{newSeat(0), newSeat(1)},
			withBattlefield(
				creature(cardID(10), 0, "Captain", 3, 3, opts...),
				creature(cardID(20), 1, "Ogre", 3, 3),
			),
			withTurn(7, 0, "declare_attackers"),
		)
	}
	plain := input(0, board(), passMove(0), attackMove(t, 0, cardID(10), 1))
	if ok, got := attacksIn(t, heuristic.New(), plain); !ok {
		t.Fatalf("an even trade with a plain 3/3 should still be taken: %s", got)
	}
	cmdr := input(0, board(commander()), passMove(0), attackMove(t, 0, cardID(10), 1))
	if ok, got := attacksIn(t, heuristic.New(), cmdr); ok {
		t.Fatalf("attacked: %s; the commander's death costs its next cast's tax", got)
	}
	off := withConfig(func(c *heuristic.Config) { c.GangAwareAttacks = false })
	if ok, got := attacksIn(t, off, cmdr); !ok {
		t.Fatalf("with GangAwareAttacks off the commander attacked as before; got %s", got)
	}
}

// #2676, the issue's window: a 3/3 and a 2/4 attack a bot on 23 life
// with six untapped 0/1 Plants. A Plant can only ever block, so a chump
// with one spends it on its only use, and the bot blocks the 3/3.
func TestPlantsChumpBlockBecauseBlockingIsAllTheyDo(t *testing.T) {
	cards := []protocol.CardView{
		creature(cardID(20), 1, "Soldier Captain", 3, 3, tapped(), attacking(0)),
		creature(cardID(21), 1, "Y'shtola", 2, 4, tapped(), attacking(0)),
	}
	var moves []legal.Move
	for i := 10; i < 16; i++ {
		cards = append(cards, plant(cardID(i), 0))
		moves = append(moves, blockMove(t, 0, cardID(i), cardID(20)), blockMove(t, 0, cardID(i), cardID(21)))
	}
	moves = append(moves, finishBlocksMove(0))
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(23)), newSeat(1)},
		withBattlefield(cards...),
		withTurn(8, 1, "declare_blockers"),
	)
	in := input(0, v, moves...)
	d := decide(t, heuristic.New(), in)
	if d.Index < 0 || moves[d.Index].Kind != legal.KindBlock {
		t.Fatalf("chose %q; a Plant should chump the 3/3", chose(t, in, d))
	}
	if got := decodeBlock(t, moves[d.Index]); got != cardID(20) {
		t.Fatalf("blocked %s first; the 3/3 saves the most damage", got)
	}

	off := withConfig(func(c *heuristic.Config) { c.Weights.BlockOnlyBody = 0 })
	if d := decide(t, off, in); d.Index >= 0 && moves[d.Index].Kind == legal.KindBlock {
		t.Fatalf("with BlockOnlyBody off the bot blocked (%q); it took the hit before", chose(t, in, d))
	}
}

// #2676: a creature card keeps its body price. The same window with six
// nontoken 0/1s is not worth a chump at 23 life.
func TestANontokenZeroOneStillDoesNotChumpAtHealthyLife(t *testing.T) {
	cards := []protocol.CardView{
		creature(cardID(20), 1, "Soldier Captain", 3, 3, tapped(), attacking(0)),
	}
	var moves []legal.Move
	for i := 10; i < 16; i++ {
		cards = append(cards, creature(cardID(i), 0, "Wall", 0, 1))
		moves = append(moves, blockMove(t, 0, cardID(i), cardID(20)))
	}
	moves = append(moves, finishBlocksMove(0))
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(23)), newSeat(1)},
		withBattlefield(cards...),
		withTurn(8, 1, "declare_blockers"),
	)
	in := input(0, v, moves...)
	if d := decide(t, heuristic.New(), in); d.Index >= 0 && moves[d.Index].Kind == legal.KindBlock {
		t.Fatalf("chose %q; a creature card is not a blocking-only body", chose(t, in, d))
	}
}

func decodeBlock(t *testing.T, m legal.Move) string {
	t.Helper()
	var p struct {
		Attacker string `json:"attacker"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatal(err)
	}
	return p.Attacker
}

package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// assigns_no_combat_damage_test.go pins ADR 0187 §4 (#3569) at the bot:
// a creature that assigns no combat damage this turn is priced as the
// 0-power attacker it is (combatPower), wherever the block and attack
// scorers read Power as "damage it deals".

// noDamage marks a creature as assigning no combat damage this turn, the
// way the wire would (CardView.AssignsNoCombatDamage).
func noDamage() cardOpt {
	return func(c *protocol.CardView) { c.AssignsNoCombatDamage = true }
}

// blockDecision is what the bot does against one big attacker with a Bear
// and a spare 1/1 to offer: the block it chose, or "" for declining.
func blockDecision(t *testing.T, life int, attacker protocol.CardView, mine ...protocol.CardView) string {
	t.Helper()
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(life)), newSeat(1)},
		withBattlefield(append([]protocol.CardView{attacker}, mine...)...),
		withTurn(5, 1, "declare_blockers"),
	)
	moves := make([]legal.Move, 0, len(mine))
	for _, b := range mine {
		moves = append(moves, blockMove(t, 0, b.InstanceID, attacker.InstanceID))
	}
	in := input(0, v, moves...)
	d := decide(t, heuristic.New(), in)
	if d.Index == aiseat.Decline {
		return ""
	}
	return chose(t, in, d)
}

// Against a creature that assigns nothing there is no damage to chump
// away: the bot at a life total a 5/5 would threaten keeps its creatures
// home, and chumps the same 5/5 once it has no mark.
func TestBotDoesNotChumpACreatureThatAssignsNoCombatDamage(t *testing.T) {
	spare := creature(cardID(11), 0, "Spare", 1, 1)
	bear := creature(cardID(10), 0, "Bear", 2, 2)

	control := creature(cardID(20), 1, "Giant", 5, 5, attacking(0))
	if got := blockDecision(t, 4, control, bear, spare); got != "Block with "+cardID(11) {
		t.Fatalf("control (no mark): chose %q, want a chump with the spare", got)
	}
	marked := creature(cardID(20), 1, "Giant", 5, 5, attacking(0), noDamage())
	if got := blockDecision(t, 4, marked, bear, spare); got != "" {
		t.Fatalf("marked attacker: chose %q, want to decline: it will deal no damage", got)
	}
}

// A swing the opponent cannot survive is the bot's to take. A marked
// attacker's 2 damage are never assigned, so the same swing is no longer
// lethal and is no longer the all-in kill.
func TestBotDoesNotCountAMarkedAttackerTowardLethal(t *testing.T) {
	swing := func(opts ...cardOpt) (string, string) {
		v := newView(
			[]protocol.PlayerView{newSeat(0, withLife(12)), newSeat(1, withLife(2))},
			withBattlefield(creature(cardID(10), 0, "Bear", 2, 2, opts...)),
			withTurn(9, 0, "declare_attackers"),
		)
		in := input(0, v, passMove(0), attackMove(t, 0, cardID(10), 1))
		d := decide(t, heuristic.New(), in)
		return chose(t, in, d), d.Reason
	}
	got, reason := swing()
	if got == "Pass priority" || !strings.Contains(reason, "all-in for the kill") {
		t.Fatalf("control (no mark): chose %q (%s), want the lethal swing", got, reason)
	}
	_, reason = swing(noDamage())
	if strings.Contains(reason, "all-in for the kill") {
		t.Fatalf("marked attacker: %q, want no all-in kill from an attacker that assigns nothing", reason)
	}
}

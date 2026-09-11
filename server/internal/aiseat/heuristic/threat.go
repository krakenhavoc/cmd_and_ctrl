package heuristic

import "sort"

// threat.go ranks the three opponents and keeps the aggression
// rotation that stops a four-player bot from pounding one seat all
// game for no result.
//
// Commander is not a duel. "Who do I attack" is a question a heads-up
// heuristic never has to ask, and answering it badly is the single
// most visible way a multiplayer bot looks broken: it picks seat 1,
// swings into their blockers every turn, never connects, and three
// players watch it do nothing for twenty turns. The rotation below is
// the fix, and it is deliberately simple — count the turns that made
// no difference, and when there have been three of them, go and
// bother somebody else for a while.

const (
	// ineffectiveTurns is how many consecutive attack turns against
	// one seat may fail to move their life total before the bot
	// gives up on that seat. Three, per the sprint's exit criterion.
	ineffectiveTurns = 3
	// rotationCooldown is how many turns a seat stays off the target
	// list after the bot rotates away from it.
	rotationCooldown = 3
)

// rankOpponents orders the live opponents by Threat, highest first,
// with the seat index as a deterministic tiebreak. Map iteration
// order must never reach a decision.
func (w Weights) rankOpponents(evals map[string]*SeatEval, me string, order []string) []*SeatEval {
	out := make([]*SeatEval, 0, len(order))
	for _, id := range order {
		e := evals[id]
		if e == nil || e.ID == me || e.Eliminated {
			continue
		}
		out = append(out, e)
	}
	sort.SliceStable(out, func(i, j int) bool {
		ti, tj := w.Threat(out[i]), w.Threat(out[j])
		if ti != tj {
			return ti > tj
		}
		return out[i].Seat < out[j].Seat
	})
	return out
}

// aggression is the rotation's state. One per bot seat; the policy
// owns it under its mutex.
type aggression struct {
	// focus is the seat the bot is currently pressuring.
	focus string
	// ineffective counts consecutive attack turns against focus that
	// left their life total where it was.
	ineffective int
	// turn / life record the turn the last attack was declared on and
	// the focus's life at that moment, so the NEXT turn can judge
	// whether it accomplished anything.
	turn     int
	life     int
	attacked bool
	// avoid maps a seat to the turn number after which it becomes a
	// legal focus again.
	avoid map[string]int
}

// pick chooses this turn's attack target from the threat-ordered
// candidates, settling the previous turn's result first.
//
// The contract with the caller is that pick is called once per
// declare-attackers decision; it is idempotent within a turn (the
// second call on turn N does not re-settle or re-record), because the
// runner asks the policy once per attacker declared and a four-
// creature swing is four calls.
func (a *aggression) pick(turn int, cands []*SeatEval) string {
	if len(cands) == 0 {
		return ""
	}
	live := make(map[string]*SeatEval, len(cands))
	for _, c := range cands {
		live[c.ID] = c
	}

	// Settle: did last turn's attack on the focus move their life?
	if a.attacked && a.focus != "" && turn > a.turn {
		if e := live[a.focus]; e == nil || e.Life >= a.life {
			a.ineffective++
		} else {
			a.ineffective = 0
		}
		a.attacked = false
	}

	// Rotate: three ineffective turns and the seat goes on the
	// cooldown list.
	if a.focus != "" && a.ineffective >= ineffectiveTurns {
		if a.avoid == nil {
			a.avoid = make(map[string]int, 3)
		}
		a.avoid[a.focus] = turn + rotationCooldown
		a.focus, a.ineffective = "", 0
	}

	// Highest threat that is not on cooldown; if every opponent is,
	// the cooldown yields and the highest threat wins.
	next := ""
	for _, c := range cands {
		if a.avoid[c.ID] > turn {
			continue
		}
		next = c.ID
		break
	}
	if next == "" {
		next = cands[0].ID
	}
	if next != a.focus {
		a.focus, a.ineffective = next, 0
	}

	if !a.attacked || a.turn != turn {
		a.turn, a.attacked = turn, true
		if e := live[a.focus]; e != nil {
			a.life = e.Life
		}
	}
	return a.focus
}

// reset drops the rotation state. Used when the bot's seat changes
// hands in a test; the policy is otherwise stateless between games
// because a fresh one is constructed per seat.
func (a *aggression) reset() { *a = aggression{} }

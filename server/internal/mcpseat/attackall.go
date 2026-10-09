package mcpseat

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// The declare-attackers window, made explicit for the model (#2793).
//
// The enumerator offers an attack one creature at a time, and the active
// seat's pass_priority is what ends the declaration (#2462). That works
// for a bot, which re-enumerates in-process, but an agent paid an `act`
// and a `wait_for_decision` per attacker, and nothing in the list said
// that move 0, "Pass priority", was the one that committed the attack.
// So in the seat's own declare-attackers window:
//
//   - the pass is relabelled for what it does there: "No attack" before
//     anything is declared, "Done declaring attackers" after;
//   - one "attack with all" move is added per opponent that two or more
//     creatures may attack. It is the browser's #318 action: every
//     creature at ONE named seat, sent as one `declare_attackers`.
//
// The added move is composed only of pairs the enumerator itself offered
// in this window, so it lets the agent do nothing the list does not
// already allow (ADR 0122, owner decision 1); it only saves the round
// trips. It lives here and not in the shared enumerator so the bots,
// which pick from that list, are unchanged (ADR 0033 §1).

// typeDeclareAttackers is the bulk verb's wire value (docs/protocol.md,
// #318). The enumerator never offers it, so legal has no constant, and
// this package may not import actions (imports_test.go).
const typeDeclareAttackers = "declare_attackers"

// stepDeclareAttackers is game.StepDeclareAttackers' wire value.
const stepDeclareAttackers = "declare_attackers"

// attackMoveParams is the part of a declare_attacker move's params this
// file reads: legal's attackParams.
type attackMoveParams struct {
	Attacker      string `json:"attacker"`
	Target        string `json:"target"`
	PhyrexianLife int    `json:"phyrexian_life,omitempty"`
	Exert         bool   `json:"exert,omitempty"`
}

type attackAllEntry struct {
	Attacker string `json:"attacker"`
	Target   string `json:"target"`
}

// attackAllPayload is declare_attackers' params: the set, and the
// auto_tap the browser sends with it so a tax the lands can pay is paid
// (ADR 0080).
type attackAllPayload struct {
	Attackers []attackAllEntry `json:"attackers"`
	AutoTap   bool             `json:"auto_tap,omitempty"`
}

// declaringAttackers reports whether the view is the seat's own
// declare-attackers step: its pass there ends the declaration.
func declaringAttackers(v *protocol.GameView, me string) bool {
	if v == nil || v.Turn.Step != stepDeclareAttackers {
		return false
	}
	self := mySeat(v, me)
	return self != nil && v.Turn.ActiveSeat == self.Seat
}

// withAttackDeclaration is the window's list with the declare-attackers
// step spelled out, as above. Outside that step it is the list
// unchanged. Server moves keep their numbers; the added moves go on the
// end. Calling it on a list it already extended changes nothing.
func withAttackDeclaration(v *protocol.GameView, me string, moves []legal.Move, partial bool) []legal.Move {
	if !declaringAttackers(v, me) {
		return moves
	}
	for _, m := range moves {
		if m.Type == typeDeclareAttackers {
			return moves
		}
	}
	out := append([]legal.Move(nil), moves...)
	attacking := 0
	for i := range v.Battlefield.Cards {
		c := &v.Battlefield.Cards[i]
		if c.Controller == me && c.AttackingTarget != "" {
			attacking++
		}
	}
	for i := range out {
		if out[i].Type == legal.TypePassPriority {
			out[i].Label = attackPassLabel(attacking)
		}
	}
	// A list the server could only send cut may be missing attackers,
	// and an "all" that is not all would be a guess.
	if partial {
		return out
	}
	return append(out, attackAllMoves(v, me, moves)...)
}

// attackPassLabel is the pass's label while the seat declares attackers.
func attackPassLabel(attacking int) string {
	switch attacking {
	case 0:
		return "No attack: pass priority without declaring attackers"
	case 1:
		return "Done declaring attackers: pass priority, attacking with 1 creature"
	default:
		return fmt.Sprintf("Done declaring attackers: pass priority, attacking with %d creatures", attacking)
	}
}

// attackAllMoves is one declare_attackers move per opposing player that
// two or more creatures may attack, in the order the list first names
// them.
func attackAllMoves(v *protocol.GameView, me string, moves []legal.Move) []legal.Move {
	type group struct {
		player    uuid.UUID
		attackers []string
		seen      map[string]bool
		lifeTax   bool
	}
	var order []string
	groups := map[string]*group{}
	exertable := map[string]bool{}
	for _, m := range moves {
		if m.Type != legal.TypeDeclareAttacker {
			continue
		}
		var p attackMoveParams
		if json.Unmarshal(m.Params, &p) != nil || p.Attacker == "" {
			continue
		}
		if p.Exert {
			// The twin that also exerts it (ADR 0130 §6). "All" sends
			// the plain attack; exerting stays a per-creature choice.
			exertable[p.Attacker] = true
			continue
		}
		// Players only, as the browser's control: a planeswalker or a
		// battle is attacked one creature at a time.
		if p.Target == me || seatByID(v, p.Target) == nil {
			continue
		}
		g := groups[p.Target]
		if g == nil {
			g = &group{player: m.Player, seen: map[string]bool{}}
			groups[p.Target] = g
			order = append(order, p.Target)
		}
		if p.PhyrexianLife > 0 {
			g.lifeTax = true
		}
		if !g.seen[p.Attacker] {
			g.seen[p.Attacker] = true
			g.attackers = append(g.attackers, p.Attacker)
		}
	}
	var out []legal.Move
	for _, target := range order {
		g := groups[target]
		n := len(g.attackers)
		// One creature already has its own move. A tax paid with life
		// (ADR 0131) is priced per creature against the life left, which
		// one bulk payment cannot answer.
		if n < 2 || g.lifeTax {
			continue
		}
		at := attackTargetFor(v, target)
		// CR 508.1c: the verb refuses a set over the limit whole.
		if at != nil && at.AttackLimit != nil && n > *at.AttackLimit {
			continue
		}
		payload := attackAllPayload{}
		for _, a := range g.attackers {
			payload.Attackers = append(payload.Attackers, attackAllEntry{Attacker: a, Target: target})
		}
		var notes []string
		if at != nil && at.Tax != "" {
			payload.AutoTap = true
			notes = append(notes, fmt.Sprintf("pays %s each, %s in all", at.Tax, strings.Repeat(at.Tax, n)))
		}
		for _, a := range g.attackers {
			if exertable[a] {
				notes = append(notes, "exerts none of them")
				break
			}
		}
		params, err := json.Marshal(payload)
		if err != nil {
			continue
		}
		label := fmt.Sprintf("Attack %s with all %d creatures that can attack them", seatByID(v, target).Name, n)
		if len(notes) > 0 {
			label += " (" + strings.Join(notes, "; ") + ")"
		}
		out = append(out, legal.Move{
			Type:   typeDeclareAttackers,
			Player: g.player,
			Kind:   legal.KindAttack,
			Label:  label,
			Params: params,
		})
	}
	return out
}

// seatByID is the seat with player id `id`, or nil.
func seatByID(v *protocol.GameView, id string) *protocol.PlayerView {
	for i := range v.Seats {
		if v.Seats[i].ID == id {
			return &v.Seats[i]
		}
	}
	return nil
}

// attackTargetFor is the turn's attack_targets row for a player, or nil.
func attackTargetFor(v *protocol.GameView, id string) *protocol.AttackTargetView {
	for i := range v.Turn.AttackTargets {
		if t := &v.Turn.AttackTargets[i]; t.Kind == "player" && t.ID == id {
			return t
		}
	}
	return nil
}

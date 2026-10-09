package heuristic

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// holdinstants.go is ADR 0136 §5 (PR 5, owner answer 6, #2668): instants
// in the turn plan, cast now or held for the end step before the bot's
// turn (Config.PlanHoldInstants).
//
// Mana the bot leaves untapped in its own turn is still there in the end
// step before its next one, because nothing untaps until its untap step
// (CR 502.3). So an instant-speed spell loses nothing by waiting for that
// step, and ADR 0126 §5 already casts it there (state.beforeMyUntap). In
// a plan, an instant-speed member is HELD unless
//
//   - it adds mana that a later member of the plan needs (§4 item 1), or
//   - it draws or tutors and there is mana left after it, so the card it
//     finds can still be cast this turn.
//
// A held member stays in the plan. It keeps its value, and its mana is
// reserved: the model pays it after every member cast this turn, so no
// other member can spend it, and what it adds arrives only for the other
// held members. When every member left is held the bot passes.
//
// A plan of one is today's choice (§3), so a lone instant-speed cast is
// made as before, unless it is what remains of a plan chosen earlier
// this turn that held it (the owner's decision of 2026-10-09, question
// 1). The policy keeps the cards each plan held for the rest of the
// turn (noteHolds, a second exception to §3's "the plan is not
// stored", beside the tax memory in tax.go), and in a later main-phase
// window that card, alone, is held again by the same rule
// (holdRemainder). Only an instant whose declared purpose is ramp or
// draw is held at all (question 2, planCandidate.instant).
//
// A member is any cast §1 allows, plus, under this switch, a cast whose
// only other cost is sacrificing lands that its own declared purpose
// more than replaces (landSacrificeIsNetMana, #2469): Harrow, the card
// #2668 is about, and the one §2 and §5 name as a member.

// planEligibleIn is planEligible, plus, with Config.PlanHoldInstants, a
// land swap that nets lands (Harrow).
func (p *Policy) planEligibleIn(st *state, m legal.Move) bool {
	if planEligible(m) {
		return true
	}
	if !p.cfg.PlanHoldInstants || m.Kind != legal.KindCast {
		return false
	}
	if c := m.Cost; c != nil {
		if c.Life > 0 || c.PhyrexianLife > 0 || c.Loyalty != 0 || len(c.Counters) > 0 || c.Hand > 0 || c.Energy > 0 || c.Exert {
			return false
		}
	}
	return p.landSacrificeIsNetMana(st, m)
}

// holdsFor is §5 for one set: which of its instant-speed members are
// held for the end step, and whether the set is payable with those
// holds. Every instant-speed member starts held. While the set cannot be
// paid, the first held member in order that adds mana is cast now
// instead, because a later member needs its mana. Then each held member
// that draws or tutors is cast now if the mana left right after it is
// more than none.
func holdsFor(base []uint8, cands []*planCandidate, order []int, mask uint, scratch []uint8) (uint, bool) {
	var held uint
	for _, i := range order {
		if bit := uint(1) << i; mask&bit != 0 && cands[i].instant {
			held |= bit
		}
	}
	if held == 0 {
		return 0, planFeasible(base, cands, order, mask, scratch)
	}
	for {
		if _, ok := planWalk(base, cands, order, mask, held, -1, scratch); ok {
			break
		}
		next := -1
		for _, i := range order {
			if held&(1<<i) != 0 && len(cands[i].adds) > 0 {
				next = i
				break
			}
		}
		if next < 0 {
			return 0, false
		}
		held &^= 1 << next
	}
	for _, i := range order {
		bit := uint(1) << i
		if held&bit == 0 || !cands[i].drawer {
			continue
		}
		if left, ok := planWalk(base, cands, order, mask, held&^bit, i, scratch); ok && left > 0 {
			held &^= bit
		}
	}
	return held, true
}

// turnHolds is the cards the plans chosen this turn held for the end
// step, by instance ID: game-independent policy state, like the
// aggression rotation, and never game state.
type turnHolds struct {
	seat string
	turn int
	ids  map[string]bool
}

// noteHolds remembers the members a plan chosen in this sorcery-speed
// window held, for the rest of the turn. A new turn forgets them.
func (p *Policy) noteHolds(st *state, moves []legal.Move, plan []aiseat.PlanMember) {
	if !p.cfg.PlanHoldInstants || !st.sorcerySpeed {
		return
	}
	if h := p.heldThisTurn; h != nil && (h.seat != st.me || h.turn != st.view.Turn.Seq) {
		p.heldThisTurn = nil
	}
	for _, pm := range plan {
		if !pm.Held || pm.Index < 0 || pm.Index >= len(moves) {
			continue
		}
		id := decode[castParams](moves[pm.Index].Params).InstanceID
		if id == "" {
			continue
		}
		if p.heldThisTurn == nil {
			p.heldThisTurn = &turnHolds{seat: st.me, turn: st.view.Turn.Seq, ids: map[string]bool{}}
		}
		p.heldThisTurn.ids[id] = true
	}
}

// holdRemainder reports whether the move decideGeneral would make,
// moves[i], casts a card a plan held earlier this turn that §5 still
// holds as a plan of one: in the bot's own main phase with an empty
// stack, a ramp or draw instant that does not draw with mana left after
// it.
func (p *Policy) holdRemainder(st *state, moves []legal.Move, vals []float64, i int) bool {
	if !p.cfg.PlanTurnMana || !p.cfg.PlanHoldInstants || !st.sorcerySpeed || i < 0 || i >= len(moves) {
		return false
	}
	h := p.heldThisTurn
	if h == nil || h.seat != st.me || h.turn != st.view.Turn.Seq {
		return false
	}
	m := moves[i]
	if !p.planEligibleIn(st, m) {
		return false
	}
	cp := decode[castParams](m.Params)
	if !h.ids[cp.InstanceID] {
		return false
	}
	card := st.castSource(cp.InstanceID)
	if card == nil {
		return false
	}
	cost, ok := st.planCastCost(m, card, cp)
	if !ok {
		return false
	}
	cands := p.rankCandidates(st, []*planCandidate{{index: i, move: m, card: card, value: vals[i], cost: cost}})
	held, _ := holdsFor(st.manaAvailable(p.cfg.PlanFilterLands), cands, []int{0}, 1, nil)
	return held&1 != 0
}

// holdDecision is the pass a held remainder makes. Its trace names the
// card as a held plan member, so a replay of the turn's windows shows
// what the policy kept.
func holdDecision(moves []legal.Move, i int, val float64) (aiseat.Decision, []aiseat.PlanMember) {
	name := moves[i].Label
	return aiseat.Decision{Index: passOrDecline(moves), Reason: fmt.Sprintf("plan: held for the end step before my turn: %s (+%.2f)", name, val)},
		[]aiseat.PlanMember{{Index: i, Label: name, Held: true}}
}

// passOrDecline is the pass on offer, or a decline when there is none
// (the runner turns a decline into a pass whenever one exists).
func passOrDecline(moves []legal.Move) int {
	if i := indexOfKind(moves, legal.KindPass); i >= 0 {
		return i
	}
	return aiseat.Decline
}

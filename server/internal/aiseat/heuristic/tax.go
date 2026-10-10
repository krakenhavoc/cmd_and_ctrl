package heuristic

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// tax.go is ADR 0136's amendment of 2026-10-09: an opponent's tax
// against the plan (Config.PlanWeighTaxes).
//
// An opponent's Rhystic Study asks "pay {1}?" when the bot casts a
// spell, and Smothering Tithe asks "pay {2}?" when it draws. Both
// arrive between two of the plan's casts: the first member's cast or
// draw raises the prompt, and the plan's next member is cast in the
// window after it. The bot used to pay every such tax it could, and in
// run 1 of PR 4b that was the largest class of plan misses (28 of 52):
// the tax spent the mana the next member needed.
//
// Now, when paying would leave the plan's next member unpayable this
// turn and the bot could pay for that member if it declined, it weighs
// what the tax prevents against what the member is worth:
//
//   - What the tax prevents is read from the prompt's source, the
//     opponent's permanent: the triggered ability rows it declares
//     (ADR 0126 §6). A row's `draws` is a card for that opponent and
//     its `tokens` a token (a Treasure), priced by giftsValue, the
//     opposition weights' share of what that opponent gains: Hand per
//     card and TokenWeight per token. No new weight.
//   - What the member is worth is its price in the window that chose
//     the plan (valueOf).
//
// The bot declines when the member is worth more. In every other case,
// including a tax whose source declares nothing, it answers as before.
//
// A tax prompt offers only pay and decline (legal.choiceMoves), so the
// window that answers it cannot rebuild the plan. The policy remembers
// the members after the first from the window that chose the plan,
// for the rest of that phase, and that memory is all this file reads.
// It is the one part of the plan that is stored (ADR 0136 §3 says
// the plan itself never is).

// planTail is what the policy remembers of the plan it chose: the seat,
// turn and phase it was chosen in, and its members in order.
type planTail struct {
	seat    string
	turn    int
	phase   int
	members []plannedCast
}

// plannedCast is one remembered member: the card, the mana its cast
// charges, what it is worth, and the mana it adds once it resolves.
type plannedCast struct {
	id    string
	name  string
	cost  manaCost
	value float64
	adds  []manaSource
	held  bool
}

// notePlan remembers the plan decideGeneral chose in a sorcery-speed
// window of the bot's own turn, or forgets the last one when this
// window chose none. Other windows leave the memory as it is: the
// windows between two of the plan's casts are the ones it is for.
func (p *Policy) notePlan(st *state, moves []legal.Move, plan []aiseat.PlanMember) {
	if !p.cfg.PlanWeighTaxes || !st.sorcerySpeed {
		return
	}
	if len(plan) < 2 {
		p.tail = nil
		return
	}
	t := &planTail{seat: st.me, turn: st.view.Turn.Seq, phase: st.view.Turn.PhaseID}
	for _, pm := range plan {
		if pm.Index < 0 || pm.Index >= len(moves) {
			p.tail = nil
			return
		}
		m := moves[pm.Index]
		cp := decode[castParams](m.Params)
		card := st.castSource(cp.InstanceID)
		if card == nil {
			p.tail = nil
			return
		}
		cost, ok := st.planCastCost(m, card, cp)
		if !ok {
			p.tail = nil
			return
		}
		v, _ := p.valueOf(st, m)
		t.members = append(t.members, plannedCast{
			id: card.InstanceID, name: card.Name, cost: cost, value: v,
			adds: castAddsMana(card, castPurpose(card, cp), p.cfg.PlanFilterLands),
			held: pm.Held,
		})
	}
	p.tail = t
}

// taxAgainstPlan answers an opponent's mana tax when paying it would
// leave the remembered plan's next member unpayable this turn. ok is
// false when the amendment has nothing to say, and the prompt is
// answered as before. decline is the answer, and why the reason.
func (p *Policy) taxAgainstPlan(st *state, ch *protocol.PendingChoiceView) (decline bool, why string, ok bool) {
	if !p.cfg.PlanWeighTaxes || ch == nil || ch.Kind != choicePayUnless || ch.Chooser != st.me ||
		ch.PayCost == "" || ch.PayCards != nil || ch.PayEnergy != nil || ch.TapCost != nil || ch.PhyrexianSymbols > 0 {
		return false, "", false
	}
	src := st.bf[ch.Source]
	if src == nil || src.Controller == st.me || src.Controller == "" {
		return false, "", false
	}
	if st.targetedByMe(src.InstanceID) {
		// Ward: the decline counters the bot's own spell, not a gift
		// the source's rows describe.
		return false, "", false
	}
	gift, ok := declineGift(src)
	if !ok {
		return false, "", false
	}
	t := p.tail
	if t == nil || t.seat != st.me || t.turn != st.view.Turn.Seq || t.phase != st.view.Turn.PhaseID {
		return false, "", false
	}
	// The next member: the first one still to cast, in hand or in the
	// command zone and not held for the end step. The members before it
	// that are still on the stack add their mana once they resolve.
	var units []uint8
	var next *plannedCast
	var pending []manaSource
	for i := range t.members {
		m := &t.members[i]
		if st.mine[m.id] != nil {
			if !m.held {
				next = m
				break
			}
			continue
		}
		if st.onStack(m.id) {
			pending = append(pending, m.adds...)
		}
	}
	if next == nil {
		return false, "", false
	}
	units = st.manaAvailable(p.cfg.PlanFilterLands)
	for _, a := range pending {
		units = addSource(units, a, next.cost.colored)
	}
	if _, payable := payMana(append([]uint8(nil), units...), next.cost, nil); !payable {
		return false, "", false
	}
	tax := parseManaCost(ch.PayCost, 0)
	rest, paid := payMana(append([]uint8(nil), units...), tax, next.cost.colored)
	if !paid {
		return false, "", false
	}
	if _, still := payMana(rest, next.cost, nil); still {
		return false, "", false
	}
	cost := -p.giftsValue(st, map[string]*giftAmounts{src.Controller: &gift}, nil)
	what := "a card"
	if gift.draws == 0 {
		what = "a token"
	}
	if next.value > cost {
		return true, fmt.Sprintf("tax: keep the mana for %s (+%.2f) over %s for its owner (−%.2f)", next.name, next.value, what, cost), true
	}
	return false, fmt.Sprintf("tax: pay it; %s for its owner (−%.2f) outweighs %s (+%.2f)", what, cost, next.name, next.value), true
}

// declineGift is what an opponent's permanent declares its triggered
// rows give its controller: the largest row's cards drawn and tokens
// made (ADR 0126 §6). False when no triggered row declares either.
func declineGift(src *protocol.CardView) (giftAmounts, bool) {
	var best giftAmounts
	found := false
	for i := range src.AbilityRows {
		r := &src.AbilityRows[i]
		if r.Kind != "triggered" || r.Purpose == nil || (r.Purpose.Draws <= 0 && r.Purpose.Tokens <= 0) {
			continue
		}
		if !found || r.Purpose.Draws+r.Purpose.Tokens > best.draws+best.tokens {
			best = giftAmounts{draws: r.Purpose.Draws, tokens: r.Purpose.Tokens}
			found = true
		}
	}
	return best, found
}

// onStack reports whether the card with this instance ID is a spell on
// the stack.
func (st *state) onStack(id string) bool {
	if st.stack[id] != nil {
		return true
	}
	for i := range st.view.StackItems {
		if st.view.StackItems[i].SourceCardID == id && st.view.StackItems[i].Kind == "spell" {
			return true
		}
	}
	return false
}

// targetedByMe reports whether a stack item the bot controls targets
// the permanent id: the shape of a ward prompt.
func (st *state) targetedByMe(id string) bool {
	for i := range st.view.StackItems {
		it := &st.view.StackItems[i]
		if it.Controller != st.me {
			continue
		}
		for _, t := range it.Targets {
			if t.ID == id {
				return true
			}
		}
	}
	return false
}

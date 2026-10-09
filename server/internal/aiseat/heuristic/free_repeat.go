package heuristic

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// free_repeat.go — #2777: a free activation is worth making once a turn.
//
// Shaman en-Kor's "{0}: The next 1 damage that would be dealt to this
// creature this turn is dealt to target creature you control instead"
// costs nothing, so the flat activation price (ActivateBase, plus the
// creature it points at) beat passing in every window, and the bot
// bought the same shield again and again. Each extra shield adds
// nothing until damage is dealt, the CR 732 breaker named the ability
// and the table paused, the shape #2449 (Equip {0}) and #2500 (a
// self-untap that nets nothing) were before it.
//
// Those two could read from the board that the move does nothing. This
// one cannot: what a free row buys is not on the wire, and the shield
// it made last time is a line of banner text. What the policy can know
// is that it already made this activation this turn. So a free
// activation — no mana, no tap, no other cost component, which is what
// makes it repeatable without limit — is priced at its usual value the
// first time in a turn, and as a no-op (below passing) after that, until
// the turn ends. Every effect a free row makes that lasts "this turn"
// is then in place, and one that is spent at once (a pump that ends at
// end of turn) was the payoff the first activation bought.
//
// It is policy memory like heldThisTurn, keyed by seat and turn, and
// written only for the move the policy chose: a row priced from the
// move list alone would loop the way it did before.

// freeRepeat prices a free activation the bot already made this turn:
// below passing, like idleEquipMove and for the same reason.
const freeRepeat = -1.0

// turnFreeActivations is the free activations the bot chose this turn,
// keyed by freeActivationKey.
type turnFreeActivations struct {
	seat string
	turn int
	keys map[string]bool
}

// freeActivationKey names one activated row of one permanent.
func freeActivationKey(cp activateParams) string {
	return cp.SourceCardID + "#" + strconv.Itoa(cp.AbilityIndex)
}

// freeActivation reports whether m activates a row of the bot's own
// permanent that costs nothing at all: no mana, no tap of the source or
// of anything else, and no other cost component.
func (p *Policy) freeActivation(st *state, m legal.Move) bool {
	if m.Kind != legal.KindActivate || !p.costsOnlyManaAndTaps(st, m) {
		return false
	}
	cp := decode[activateParams](m.Params)
	if cp.XValue > 0 || len(cp.CrewIDs) > 0 || len(cp.WaterbendIDs) > 0 || len(cp.TapIDs) > 0 {
		return false
	}
	src := st.bf[cp.SourceCardID]
	if src == nil || src.Controller != st.me {
		return false
	}
	row := rowAt(src, cp.AbilityIndex)
	if row == nil || row.TapCost || row.Exert || row.CrewCost > 0 || row.TapOthersLabel != "" ||
		row.EnergyCost > 0 || row.EnergyCostX {
		return false
	}
	if c := m.Cost; c != nil && c.Mana != "" && manaValue(c.Mana, 0) > 0 {
		return false
	}
	return manaValue(row.ManaCost, 0) == 0 && !costHasX(row.ManaCost)
}

// repeatsFreeActivation reports whether m is a free activation the bot
// already chose this turn.
func (p *Policy) repeatsFreeActivation(st *state, m legal.Move) bool {
	h := p.freeThisTurn
	if h == nil || h.seat != st.me || h.turn != st.turn {
		return false
	}
	return h.keys[freeActivationKey(decode[activateParams](m.Params))] && p.freeActivation(st, m)
}

// noteFreeActivation remembers the decision's move when it is a free
// activation. A new turn forgets the last one's.
func (p *Policy) noteFreeActivation(st *state, moves []legal.Move, d aiseat.Decision) {
	if h := p.freeThisTurn; h != nil && (h.seat != st.me || h.turn != st.turn) {
		p.freeThisTurn = nil
	}
	if d.Index < 0 || d.Index >= len(moves) || !p.freeActivation(st, moves[d.Index]) {
		return
	}
	if p.freeThisTurn == nil {
		p.freeThisTurn = &turnFreeActivations{seat: st.me, turn: st.turn, keys: map[string]bool{}}
	}
	p.freeThisTurn.keys[freeActivationKey(decode[activateParams](moves[d.Index].Params))] = true
}

// costHasX reports whether a printed mana cost has an {X}.
func costHasX(cost string) bool {
	for i := 0; i+2 < len(cost); i++ {
		if cost[i] == '{' && (cost[i+1] == 'X' || cost[i+1] == 'x') && cost[i+2] == '}' {
			return true
		}
	}
	return false
}

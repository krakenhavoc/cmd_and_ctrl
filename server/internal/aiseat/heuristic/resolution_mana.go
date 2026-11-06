package heuristic

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// resolution_mana.go — #4980: a row that adds mana on one numbered
// resolution of it in a turn, priced by the mana it returns.
//
// Soulbright Seeker's "{R}: Target creature you control gains trample
// until end of turn. If this is the third time this ability has resolved
// this turn, add {R}{R}{R}{R}." declared nothing, so it was priced as an
// undeclared targeted row, and a direction alone (#4302) would price
// every repeat of the trample at below a pass: the creature already has
// it. The third activation, the one that pays for itself and three more,
// was never priced for that mana.
//
// The row now declares the mana and which resolution adds it
// (`resolution_mana: {resolution: 3, mana: "{R}{R}{R}{R}"}`), and the
// view stamps the row with the engine's own count, `resolved_this_turn`:
// Game.ResolvedThisTurn for the row's label, the count the card's effect
// reads, per object. An activation still on the stack has not resolved
// and is not in it, so the policy adds the row's items on the stack to
// it: the move on offer would be resolution resolved + pending + 1. A
// counted resolution, not a remembered activation, is what the card
// says, and it stays true across a countered activation and a seat that
// restarted mid-turn, where the policy's per-turn memory would not.
//
// With Config.PriceResolutionMana on, an activation on the way to that
// resolution, or the one that makes it, is priced by the casts the mana
// the run returns pays for, and only when that mana comes back and can be
// used:
//
//   - the run is reachable: the mana for every activation still needed,
//     this one included, is untapped or in the pool;
//   - it nets mana: the mana added is more than the run still costs;
//   - it is the bot's own main phase, where the mana, which empties at
//     the end of the step (CR 106.4), can pay for a sorcery-speed cast;
//   - it casts more: the most mana the bot's hand and command zone could
//     spend after the run is more than they could spend without it.
//
// The price is SpellPerMana for each mana the hand can spend after the
// run, the value-per-mana proxy a cast already uses, in place of the flat
// ActivateBase: what a ritual is worth is the casts it pays for. A run
// that nets mana never takes away a cast the bot could make without it,
// since all of the bot's mana is still there and more besides, so it is
// fair to rank it against a cast now: when it wins, the casts follow. A
// repeat of a grant the bot's creature already has costs nothing more on
// such a run (the mana is the point), so neither the redundant grant nor
// the per-turn grant memory blocks it, and the row is not a mana sink
// (mana_sink.go): it makes mana rather than spending what the casts
// leave. Anywhere else, or with the switch off, the row is priced as any
// other trample grant (temporary_effect.go). The cast the mana pays for
// is priced on its own once the mana is in the pool.

// resolutionManaValue prices an activation of a row that declares the
// mana it adds on one numbered resolution this turn. False when this
// Config, the row or the window does not price it, which keeps the
// row's ordinary price.
func (p *Policy) resolutionManaValue(st *state, m legal.Move) (float64, bool) {
	if !p.cfg.PriceResolutionMana || m.Kind != legal.KindActivate || st.view == nil || st.seat == nil {
		return 0, false
	}
	if !st.myTurn || (st.step != "precombat_main" && st.step != "postcombat_main") {
		return 0, false
	}
	cp := decode[activateParams](m.Params)
	src := st.bf[cp.SourceCardID]
	if src == nil || src.Controller != st.me {
		return 0, false
	}
	row := rowAt(src, cp.AbilityIndex)
	if row == nil || row.Purpose == nil || row.Purpose.ResolutionMana == nil || row.SacrificeSelf {
		return 0, false
	}
	rm := row.Purpose.ResolutionMana
	n := row.ResolvedThisTurn + st.pendingResolutions(src, row) + 1
	if n > rm.Resolution {
		return 0, false
	}
	spend := (rm.Resolution - n + 1) * st.manaSpentBy(m)
	have := st.myMana + protocol.ManaPoolTotal(st.seat.ManaPool)
	if spend > have {
		return 0, false
	}
	net := manaValue(rm.Mana, 0) - spend
	if net <= 0 {
		return 0, false
	}
	after := st.castableManaWithin(have + net)
	if after <= st.castableManaWithin(have) {
		return 0, false
	}
	return p.cfg.SpellPerMana * float64(after), true
}

// pendingResolutions is how many items of row's ability from src are on
// the stack: activations that will resolve, and be counted, before the
// turn is over unless something stops them.
func (st *state) pendingResolutions(src *protocol.CardView, row *protocol.ActivatedAbilityView) int {
	n := 0
	for i := range st.view.StackItems {
		it := &st.view.StackItems[i]
		if it.SourceCardID == src.InstanceID && it.Label == row.Label {
			n++
		}
	}
	return n
}

// castableManaWithin is the most mana `budget` could spend on the
// non-land cards in the bot's hand and command zone (commander tax
// included), each cast at most once: a subset sum over their mana
// values. It counts mana, not colours, as biggestCastWithin does.
func (st *state) castableManaWithin(budget int) int {
	if st.seat == nil || budget <= 0 {
		return 0
	}
	can := make([]bool, budget+1)
	can[0] = true
	add := func(c *protocol.CardView, tax int) {
		if isLand(c) {
			return
		}
		mv := manaValue(c.ManaCost, 0) + tax
		if mv <= 0 || mv > budget {
			return
		}
		for b := budget; b >= mv; b-- {
			if can[b-mv] {
				can[b] = true
			}
		}
	}
	for i := range st.seat.Hand.Cards {
		add(&st.seat.Hand.Cards[i], 0)
	}
	for i := range st.seat.Command.Cards {
		c := &st.seat.Command.Cards[i]
		add(c, 2*st.seat.CommanderCasts[c.InstanceID])
	}
	for b := budget; b > 0; b-- {
		if can[b] {
			return b
		}
	}
	return 0
}

// ownPicksOnly reports whether every pick of ts is a permanent the bot
// controls: a resolution-mana run's repeated grant costs nothing more
// there, where on another player's permanent it is still a gift.
func (st *state) ownPicksOnly(ts []targetRef) bool {
	for _, t := range ts {
		c := st.bf[t.ID]
		if t.Kind == "player" || c == nil || c.Controller != st.me {
			return false
		}
	}
	return true
}

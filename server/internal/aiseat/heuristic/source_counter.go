package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// source_counter.go prices the +1/+1 counters an activated row of the
// bot's own puts on its own source, to stay (#5083): Hangarback
// Walker's "{1}, {T}: Put a +1/+1 counter on this creature."
//
// The row declared only what it answers (pump), so it was priced at the
// flat ActivateBase (0.50), less the blocker its tap costs (0.30): 0.20
// on the bot's own turn, under PassThreshold (0.25). It was taken only
// in the end step before the bot's untap, where the tap costs nothing.
// Its purpose now declares the counter (`source_counters: 1`), and
// with Config.PriceSourceCounters on (off by default, see below), each
// counter is worth what the
// evaluation counts a point of power and a point of toughness
// (Weights.Power + Weights.Toughness, 1.45), the price a landfall or
// discard payoff's counter on its source already has. It takes the
// place of ActivateBase, as a declared draw does. The counter Hangarback
// also turns into a Thopter when it dies is left out, so the price is a
// floor.
//
// The tap is priced where it always was (moves.go, tapCreatureCost): a
// blocker, and before combat on the bot's own turn the attack it gives
// up, so a Walker with an attack worth more than the counter attacks.
// On an opponent's turn before blockers are declared the counter is not
// priced and the row keeps the flat price: tapping a creature then
// takes a blocker out of the coming combat, which the flat blocker
// price does not see.
//
// A counter only grows a creature: a source that is not one of the
// bot's creatures on the battlefield is not priced either.
//
// The switch ships OFF. In #5083's arena (thopter-battle, two seats with
// the price against two without, 48 games) the seats with it won 17 of
// 96 seat-games against 31. With it on, the Walker grew before combat on
// the bot's own turn (177 of 185 activations), so it stayed tapped
// through every opponent's turn, where the flat price grows it in the
// end step before the bot's untap for nothing. The blocker price
// (0.30) undercounts a round of blocks against a counter the bot gets
// one window later anyway. `heuristic-sourcecounters` seats it for the
// next measurement.

// sourceCounterValue prices the +1/+1 counters row declares it puts on
// src, false when this Config or the window does not price them.
func (p *Policy) sourceCounterValue(st *state, src *protocol.CardView, row *protocol.ActivatedAbilityView) (float64, bool) {
	if !p.cfg.PriceSourceCounters || src == nil || row == nil || row.Purpose == nil || row.Purpose.SourceCounters <= 0 {
		return 0, false
	}
	if st.bf[src.InstanceID] == nil || src.Controller != st.me || !isCreature(src) || row.SacrificeSelf {
		return 0, false
	}
	if !st.myTurn && blockersStillToCome[st.step] {
		return 0, false
	}
	return (st.w.Power + st.w.Toughness) * float64(row.Purpose.SourceCounters), true
}

// blockersStillToCome is the set of steps of an opponent's turn in which
// a creature the bot taps is one fewer blocker for that turn's combat.
var blockersStillToCome = map[string]bool{
	"untap":             true,
	"upkeep":            true,
	"draw":              true,
	"precombat_main":    true,
	"begin_combat":      true,
	"declare_attackers": true,
}

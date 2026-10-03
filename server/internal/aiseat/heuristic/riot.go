package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// riot.go — the bot's answers to riot's and unleash's entry questions
// (ADR 0109 §10 decision 6, #1556).
//
//   - Riot (entry_riot): haste when the creature enters during its
//     controller's own turn before attackers are declared, and could
//     attack — a creature without defender — and the +1/+1 counter
//     otherwise. Haste on the turn it can swing is tempo; any other
//     time it buys nothing a counter does not beat.
//   - Unleash (an optional_replacement whose entry_keyword is
//     "unleash"): always the counter, unless it is an opponent's turn
//     and the entering creature would be the bot's only untapped
//     potential blocker, which a counter would stop blocking (CR
//     702.98a).

// choiceEntryRiot is game.PendingChoiceEntryRiot on the wire.
const choiceEntryRiot = "entry_riot"

// riotBeforeAttackers is the set of steps of the bot's own turn in which
// a creature that enters can still be declared as an attacker.
var riotBeforeAttackers = map[string]bool{
	"untap":          true,
	"upkeep":         true,
	"draw":           true,
	"precombat_main": true,
	"begin_combat":   true,
}

// enteringCard is the card an entry question is about — on the stack
// for a resolving spell, else in the bot's hand, a graveyard or exile.
func (st *state) enteringCard(id string) *protocol.CardView {
	for _, m := range []map[string]*protocol.CardView{st.stack, st.mine, st.graveyard, st.exile} {
		if c := m[id]; c != nil {
			return c
		}
	}
	return nil
}

// riotWantsHaste is decision 6's riot half.
func (st *state) riotWantsHaste(ch *protocol.PendingChoiceView) bool {
	if ch == nil || !st.myTurn || !riotBeforeAttackers[st.step] {
		return false
	}
	c := st.enteringCard(ch.Source)
	return c != nil && isCreature(c) && !hasKeyword(c, "defender")
}

// riotValue scores one answer: `apply` true is the counter.
func (st *state) riotValue(ch *protocol.PendingChoiceView, apply *bool) (float64, string) {
	counter := apply != nil && *apply
	if st.riotWantsHaste(ch) {
		if counter {
			return 0.5, "riot: the counter (it could attack now)"
		}
		return 1, "riot: haste, to attack this turn"
	}
	if counter {
		return 1, "riot: the +1/+1 counter"
	}
	return 0.5, "riot: haste (too late to attack)"
}

// unleashKeepsABlocker is decision 6's unleash half: on an opponent's
// turn, a creature entering while the bot has no other untapped creature
// is the only blocker it will have, and a counter would take that away.
func (st *state) unleashKeepsABlocker(ch *protocol.PendingChoiceView) bool {
	if ch == nil || st.myTurn {
		return false
	}
	for _, c := range st.bf {
		if c.Controller == st.me && c.InstanceID != ch.Source && isCreature(c) && !c.Tapped {
			return false
		}
	}
	return true
}

// unleashValue scores one answer to unleash's "may": `apply` true takes
// the counter.
func (st *state) unleashValue(ch *protocol.PendingChoiceView, apply *bool) (float64, string) {
	counter := apply != nil && *apply
	if st.unleashKeepsABlocker(ch) {
		if counter {
			return 0.5, "unleash: the counter (it would be the only blocker)"
		}
		return 1, "unleash: no counter, keep it able to block"
	}
	if counter {
		return 1, "unleash: the +1/+1 counter"
	}
	return 0.5, "unleash: decline"
}

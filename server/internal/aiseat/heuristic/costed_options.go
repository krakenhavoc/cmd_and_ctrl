package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// costed_options.go — an option_pick whose options cost mana (#2854):
// Winter's Chill's "its controller may pay {1} or {2}", Lim-Dûl's
// Hex's "unless they pay {B} or {3}".
//
// The free option comes first on every such prompt (the queue refuses
// a cost on it), and it is the branch the card hangs its penalty on:
// the creature is destroyed, the player takes the damage, the spell is
// countered. Every costed option is one the enumerator has just found
// payable. So the rule is: pay, and pay as little as buys a way out —
// the cheapest option by mana value, which keeps the most mana for
// the rest of the turn. On Winter's Chill that is the {1} branch (the
// creature survives and deals no combat damage) rather than the {2}
// one (it fights normally): both keep the creature, and a seat whose
// creature is worth more in combat than {1} of mana is a judgement the
// wire gives the policy nothing to make.

const choiceOptionPick = "option_pick"

// costedOptionValue scores one answer to an option_pick with at least
// one costed option. The second return is false for an option pick
// whose options are all free (Torment of Hailfire, a pile split),
// which keeps the policy's existing answer: the enumerator's first.
func costedOptionValue(ch *protocol.PendingChoiceView, index *int) (float64, string, bool) {
	if ch == nil || index == nil || *index < 0 || *index >= len(ch.PickOptions) {
		return 0, "", false
	}
	costed := false
	for _, o := range ch.PickOptions {
		if o.ManaCost != "" {
			costed = true
			break
		}
	}
	if !costed {
		return 0, "", false
	}
	opt := ch.PickOptions[*index]
	if opt.ManaCost == "" {
		return 0.5, "pay nothing: " + opt.Label, true
	}
	mv := float64(manaValue(opt.ManaCost, 0))
	return 1 - 0.4*mv/(1+mv), "pay the least that buys a way out: " + opt.Label, true
}

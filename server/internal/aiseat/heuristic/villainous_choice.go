package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// villainous_choice.go — "faces a villainous choice" (CR 701.55, #3140):
// an option_pick of two options, both bad, addressed to the facing
// player. The enumerator offers both (CR 701.55b: an impossible option
// may be chosen), each with the life it costs as MoveCost.Life; the
// card puts one more hint on each option on the wire, `harm`, what the
// option costs the chooser besides life and mana, in cards (a card
// from hand is 1, a creature 2).
//
// The rule is the cheapest loss: price each option in the units
// Strength is measured in, the life through LifeCostValue (so the cost
// climbs steeply near death, the way every other life payment's does)
// and the harm as that many cards in hand, and take the smaller. An
// option that would take the bot's last life is the worst answer there
// is. An option the card marked as doing nothing to this player arrives
// with both numbers at zero and so is the cheapest, which is right: the
// rules let a player pick an option they cannot carry out and lose
// nothing.
//
// Only a prompt whose options carry `harm` is read here. A card that
// gives no hint, or one where every option is free, keeps the
// enumerator's first answer, as any other option pick does.

// villainousValue scores answering a villainous choice `ch` with
// offered option `index`. The second return is false for an option pick
// that is not one.
func (st *state) villainousValue(ch *protocol.PendingChoiceView, index *int) (float64, string, bool) {
	if ch == nil || index == nil || *index < 0 || *index >= len(ch.PickOptions) {
		return 0, "", false
	}
	hinted := false
	for _, o := range ch.PickOptions {
		if o.Harm > 0 {
			hinted = true
			break
		}
	}
	if !hinted {
		return 0, "", false
	}
	opt := ch.PickOptions[*index]
	life := st.myLife()
	if opt.LifeCost > 0 && life-opt.LifeCost <= 0 {
		return suicideValue, "villainous choice: would lose its last life: " + opt.Label, true
	}
	cost := st.w.LifeCostValue(life, opt.LifeCost) + st.w.Hand*float64(opt.Harm)
	return -cost, "villainous choice: the cheapest loss: " + opt.Label, true
}

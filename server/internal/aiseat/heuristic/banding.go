package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// banding.go — the bot and banding (CR 702.22, ADR 0224 §6, #5723).
//
// The policy is deliberately small:
//
//   - The bot never forms a band. The enumerator offers per-creature
//     attacks, which announce no band, and nothing here asks for one.
//   - The bot blocks a creature in a band only to survive. Blocking one
//     member blocks the whole band (CR 702.22h): the blocker takes every
//     member's damage, and the attacking player divides the blocker's
//     damage among them as they like (CR 702.22k), so the block scorer's
//     one-attacker trade is the wrong picture of it. What a block on a
//     band does do well is stop all of its damage at once, which is
//     exactly what a chump to survive wants, and that is priced in full.
//   - A damage division banding hands the bot (CR 702.22j-k) is answered
//     with the enumerator's protective split, which loses the bot the
//     least (legal.splitPlan.protective), through the ordinary canonical
//     first choice.

// bandMates is every other attacker in `atk`'s band, from the view's
// attack_band numbers. Nil for an attacker in no band.
func bandMates(st *state, atk *protocol.CardView) []*protocol.CardView {
	if atk == nil || atk.AttackBand == 0 {
		return nil
	}
	var out []*protocol.CardView
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.InstanceID != atk.InstanceID && c.AttackBand == atk.AttackBand && c.AttackingTarget == atk.AttackingTarget {
			out = append(out, c)
		}
	}
	return out
}

// bandDamageSaved is the damage a block on `atk` stops from its
// band-mates: what each of them connects with now, since blocking one
// member blocks them all and nothing they deal reaches the player
// unless it tramples over the blocker.
func bandDamageSaved(st *state, atk *protocol.CardView) int {
	saved := 0
	for _, m := range bandMates(st, atk) {
		saved += connects(m, blockersOn(st, m.InstanceID))
	}
	return saved
}

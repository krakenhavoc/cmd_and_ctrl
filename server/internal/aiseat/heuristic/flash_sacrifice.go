package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// flash_sacrifice.go — #3712: a permanent that may be cast "as though it
// had flash" at the price of a sacrifice at the next cleanup step
// (Necromancy, Lightning Reflexes, Parapet, ...).
//
// Cast at instant speed, such a permanent is gone at the end of the turn,
// so the cast is worth only what the permanent does before then. Cast in
// the bot's own main phase with an empty stack, it stays. The wire says
// which cards print the rider (purpose.flash_cleanup_sacrifice) and the
// move says when: any cast outside the bot's own sorcery-speed window is
// the sacrificed kind.
//
// The bot makes the instant-speed cast only when the permanent does
// something before cleanup that it can see:
//
//   - an answer: an Aura aimed at one of the bot's own permanents while
//     something is on the stack (a shroud or protection Aura in response,
//     a buff that saves the creature);
//   - a combat trick: an Aura on one of the bot's own creatures that is
//     attacking or blocking, once attackers (or blockers) are declared;
//   - a blocker: a rider permanent that is not an Aura, on an opponent's
//     turn at declare attackers (Necromancy reanimates one before blocks),
//     or at declare blockers with one of the bot's creatures in combat
//     (Parapet's toughness).
//
// Otherwise the cast waits for the bot's own main phase.

// flashSacrificeDeclined is the price of a cast that would only lose the
// permanent at cleanup. Below nothing, so a pass always beats it.
const flashSacrificeDeclined = -1.0

// printsFlashSacrifice reports whether the card prints the rider.
func printsFlashSacrifice(c *protocol.CardView) bool {
	return c != nil && c.Purpose != nil && c.Purpose.FlashCleanupSacrifice
}

// holdsFlashSacrifice reports whether the cast of card (with params cp)
// should wait for the bot's own main phase.
func (p *Policy) holdsFlashSacrifice(st *state, card *protocol.CardView, cp castParams) bool {
	if !p.cfg.HoldFlashSacrifice || !printsFlashSacrifice(card) || st.sorcerySpeed {
		return false
	}
	return !st.flashSacrificeEarnsItsKeep(card, cp)
}

// inCombat reports whether c is attacking or blocking.
func inCombat(c *protocol.CardView) bool {
	return c.AttackingTarget != "" || c.BlockingTarget != "" || len(c.BlockingTargets) > 0
}

// flashSacrificeEarnsItsKeep is the three cases above.
func (st *state) flashSacrificeEarnsItsKeep(card *protocol.CardView, cp castParams) bool {
	v := st.view
	stackEmpty := len(v.Stack.Cards) == 0 && len(v.StackItems) == 0
	combat := st.step == "declare_attackers" || st.step == "declare_blockers"
	if isType(card, "aura") {
		for _, t := range cp.Targets {
			host := st.bf[t.ID]
			if t.Kind != "card" || host == nil || host.Controller != st.me {
				continue
			}
			if !stackEmpty || (combat && inCombat(host)) {
				return true
			}
		}
		return false
	}
	if !combat {
		return false
	}
	if st.step == "declare_attackers" && !st.myTurn {
		return true
	}
	if st.step != "declare_blockers" {
		return false
	}
	for _, c := range st.bf {
		if c.Controller == st.me && inCombat(c) {
			return true
		}
	}
	return false
}

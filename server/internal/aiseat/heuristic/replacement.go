package heuristic

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// replacement.go — the bot's answers to the "may" replacement effects
// (#2390). Until #2390 every one of them took the yes the trigger
// prompt's branch gives a "you may" the bot controls, and the copy
// prompt took no answer at all. The audit, prompt by prompt:
//
// optional_replacement
//   - dredge (dredge.go): priced against the draw it replaces.
//   - unleash (riot.go): the counter, unless it costs the only blocker.
//   - CR 903.9b, a commander headed for a HAND (playable_from_zone):
//     no. In hand it is cast for its printed cost; from the command
//     zone it pays the CR 903.8 tax. ADR 0115's owner decision 2 for
//     the graveyard question, pointed at the one zone a commander can
//     always be cast from.
//   - CR 903.9b, a commander headed for a LIBRARY, and #1397's question
//     asked before a cost is paid: yes. A library is where a commander
//     is lost; the cost path's frame does not say which zone it is
//     going to, and the command zone is never worse than a library.
//   - Library of Leng's "put the discarded card on top of your library
//     instead": yes. The card the bot is discarding is not on the
//     prompt — naming a card in a hand there would hand the table the
//     correlation handle PR #513 removed — so the bot cannot price it
//     against the draw it would displace. It keeps the card: a card
//     drawn again is a card, a card in a graveyard usually is not.
//   - Moonlit Meditation's "create copies of the enchanted permanent
//     instead": yes. It enchants a permanent its controller controls,
//     which the bot chose, and the replacement is the Aura's only text.
//
// entry_pay_life (shocklands, the 3-life modal lands): pay only when the
// untapped mana buys something this turn and the life is affordable.
//
// copy_target (Clone, Phyrexian Metamorph, Spark Double, …): copy the
// most valuable permanent on offer. Declining was the enumerator's first
// answer, and the bot took it: a Clone that copies nothing enters as a
// 0/0 and dies.
//
// Mox Diamond's discard, the reveal lands, the sacrifice lands and
// devour are entry_discard_from_hand / entry_reveal_from_hand /
// entry_sacrifice, priced in choices.go, and the default yes never
// reached them.

// choiceCopyTarget is game.PendingChoiceCopyTarget on the wire.
const choiceCopyTarget = "copy_target"

// optionalReplacementValue scores one answer to an optional_replacement
// prompt. `apply` true takes the replacement.
func (p *Policy) optionalReplacementValue(st *state, ch *protocol.PendingChoiceView, apply *bool) (float64, string) {
	yes := apply != nil && *apply
	switch {
	case ch == nil:
	case ch.EntryKeyword == "unleash":
		return st.unleashValue(ch, apply)
	case ch.Dredge > 0:
		return p.dredgeValue(st, ch, yes)
	case ch.PlayableFromZone:
		if yes {
			return 0.5, "commander: headed for a hand, where it is cast without the tax"
		}
		return 1, "commander: keep it in hand"
	}
	if yes {
		return 1, "yes"
	}
	return 0.5, "no"
}

// entryPayLifeValue scores one answer to "as this enters, you may pay N
// life. If you don't, it enters tapped." Entering tapped is free and is
// the baseline, zero. Paying is worth the untapped mana when the bot
// can spend it before the land would untap anyway, less the life at
// the price LifeCostValue puts on it, which climbs steeply near death.
//
// The life comes off the move (#547's MoveCost, which legal attaches
// for this prompt since #2390). A pay branch with no price on it is
// declined: a bot never takes a life payment it cannot see.
func (p *Policy) entryPayLifeValue(st *state, m legal.Move, apply *bool) (float64, string) {
	if apply == nil || !*apply {
		return 0, "enter tapped"
	}
	if m.Cost == nil || m.Cost.Life <= 0 {
		return -1, "pay life: no price on the move"
	}
	life := st.myLife()
	if life-m.Cost.Life < p.cfg.LifeFloor {
		return suicideValue, "pay life: would pay its last life"
	}
	if !st.untappedManaBuysSomething() {
		return -st.w.LifeCostValue(life, m.Cost.Life), "pay life: the mana would go unused, enter tapped"
	}
	gain := st.w.ManaSource - st.w.TappedManaSource + p.cfg.SpellPerMana
	return gain - st.w.LifeCostValue(life, m.Cost.Life), "pay life: the mana is spent this turn"
}

// untappedManaBuysSomething reports whether one more untapped mana
// source, now, lets the bot cast more than it could without it before
// that source would untap on its own: the cards it could cast with the
// extra mana add up to more than the mana it already has.
//
// On the bot's own turn before its end step that is every spell in its
// hand and its commander (tax included). Any other time a tapped land
// untaps in the bot's next untap step, so only what it can cast before
// then counts: instants and flash.
func (st *state) untappedManaBuysSomething() bool {
	if st.seat == nil {
		return false
	}
	budget := st.myMana + len(st.seat.ManaPool)
	ownMain := st.myTurn && st.step != "end" && st.step != "cleanup"
	var spend int
	consider := func(c *protocol.CardView, tax int) {
		if isLand(c) {
			return
		}
		if !ownMain && !isType(c, "instant") && !hasKeyword(c, "flash") {
			return
		}
		if mv := manaValue(c.ManaCost, 0) + tax; mv <= budget+1 {
			spend += mv
		}
	}
	for i := range st.seat.Hand.Cards {
		consider(&st.seat.Hand.Cards[i], 0)
	}
	for i := range st.seat.Command.Cards {
		c := &st.seat.Command.Cards[i]
		consider(c, 2*st.seat.CommanderCasts[c.InstanceID])
	}
	return spend > budget
}

// copyTargetValue scores one answer to "you may have this enter as a
// copy of …" (CR 614.1c, 707.2): the permanent the copy would be, as
// the bot's own, entering untapped and (a creature) summoning-sick. A
// copy takes the copiable values and nothing else — no counters, no
// Auras, no damage — but the view carries only what the permanent is
// now, so a pumped creature is over-priced a little.
//
// Copying a legendary permanent the bot controls is worth almost
// nothing: the legend rule (CR 704.5j) keeps one of the two. A card
// whose copy is not legendary or not the same name (Spark Double,
// Sakashima) says so only in its text, so the bot passes those up too
// — weaker than the card, never a dead copy.
//
// Copying nothing scores zero, so any real permanent is preferred.
func (st *state) copyTargetValue(ids []string, lookup func(string) *protocol.CardView) (float64, string) {
	if len(ids) == 0 {
		return 0, "copy: nothing"
	}
	c := lookup(ids[0])
	if c == nil {
		return 0, "copy: a permanent it cannot see"
	}
	if st.legendaryIControl(c) {
		return 0.05, "copy: a legend it already has"
	}
	fresh := *c
	fresh.Tapped, fresh.SummoningSick, fresh.Restrictions, fresh.AttachedTo = false, false, nil, nil
	v := st.w.permanentValue(&fresh)
	if isCreature(&fresh) {
		v *= st.w.SickCreature
	}
	return v, "copy: the most valuable permanent"
}

// legendaryIControl reports whether c is legendary and the bot controls
// a permanent with its name — c itself included.
func (st *state) legendaryIControl(c *protocol.CardView) bool {
	if !strings.Contains(strings.ToLower(c.TypeLine), "legendary") {
		return false
	}
	for i := range st.view.Battlefield.Cards {
		b := &st.view.Battlefield.Cards[i]
		if b.Controller == st.me && b.Name == c.Name {
			return true
		}
	}
	return false
}

package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// discard_by_opponent.go — ADR 0153 §5: a discard an opponent forces is
// priced by what the card does when it is discarded that way.
//
// The Obstinate Baloth family puts the card onto the battlefield instead
// of the graveyard, so when Mind Rot makes the bot discard two, a Baloth
// in hand is not a loss: it is a free 4/4. A card whose own trigger pays
// (Metrognome's four Gnomes) is cheaper by what it pays.
//
// The bot reads two things off the wire, and infers neither from a name
// (ADR 0126):
//
//   - the hand card's `discard_by_opponent`, which the card's catalog
//     constructor declared: "battlefield", "library" or "trigger";
//   - the prompt's `source`, the card whose spell or ability asked, and
//     the controller of that card's stack item (or, off the stack, of the
//     card itself). A discard the bot chooses for its own loot, a cost,
//     or the cleanup step is not an opponent's, so none of this applies.
//
// The heuristic does not try to BAIT these discards (it does not hold a
// Baloth in hand on purpose); that is a strategy question for a later
// bot change. And "library" is declared but not priced here: ADR 0153 §5
// prices the battlefield and trigger cases only.

// opponentForcedDiscard reports whether ch is a prompt in which a spell
// or ability an opponent controls is making the bot discard.
func (st *state) opponentForcedDiscard(ch *protocol.PendingChoiceView) bool {
	if ch == nil || st.view == nil || ch.FromPlayer != st.me || ch.Chooser != st.me || ch.Source == "" {
		return false
	}
	controller, ok := st.controllerOfSource(ch.Source)
	return ok && controller != st.me
}

// controllerOfSource is who controls the spell or ability a card put on
// the stack, or, with none there, the permanent or spell card itself. The
// stack item is the answer CR 113.8 gives (a stolen permanent's ability
// is its thief's), so it is read first. An ability's source is the
// permanent, and a resolving spell has left the stack for its graveyard
// by the time its prompt is open.
func (st *state) controllerOfSource(id string) (string, bool) {
	for i := range st.view.StackItems {
		if it := &st.view.StackItems[i]; it.SourceCardID == id {
			return it.Controller, true
		}
	}
	if c := st.bf[id]; c != nil {
		return c.Controller, true
	}
	if c := st.stack[id]; c != nil {
		return c.Controller, true
	}
	// A resolving spell has left the stack by the time its prompt is
	// open and sits in its graveyard, still controlled by its caster.
	if c := st.graveyard[id]; c != nil {
		return c.Controller, true
	}
	return "", false
}

// forcedDiscardCredit is what discarding the bot's own card c gives back
// when an opponent is forcing it (`forced`), beyond what the bot's
// battlefield pays for any discard (discardPayoff). The discard then
// costs
//
//	handKeepValue(c) - discardPayoff(c) - forcedDiscardCredit(c)
//
// For a "battlefield" card that is -(cardValue(c)) - discardPayoff(c):
// the card comes back as a permanent, so pitching it gains what it is
// worth. For a "trigger" card it is its keep value less what its own
// trigger pays, in discardPayoff's units. Zero for every other card, and
// whenever the bot chose the discard itself.
func (st *state) forcedDiscardCredit(cfg Config, c *protocol.CardView, forced bool) float64 {
	if !forced || c == nil {
		return 0
	}
	// The declaration rides on the card in the bot's hand; a prompt's
	// option list is a separate projection of the same card.
	h := c
	if m := st.mine[c.InstanceID]; m != nil {
		h = m
	}
	switch h.DiscardByOpponent {
	case "battlefield":
		return st.handKeepValue(cfg, c) + st.cardValue(cfg, c)
	case "trigger":
		return st.ownDiscardTriggerPayoff(cfg, h)
	}
	return 0
}

// ownDiscardTriggerPayoff is what the card's own "an opponent makes you
// discard this card" trigger pays, in discardPayoff's units: a token at
// TokenWeight, a +1/+1 counter at what losing one costs, a point of
// damage to each live opponent at DamageToOpponent. Zero when the card
// declares nothing.
func (st *state) ownDiscardTriggerPayoff(cfg Config, c *protocol.CardView) float64 {
	var v float64
	for i := range c.AbilityRows {
		r := &c.AbilityRows[i]
		if r.Kind != "triggered" || r.Purpose == nil || r.Purpose.OnOpponentDiscard == nil {
			continue
		}
		d := r.Purpose.OnOpponentDiscard
		v += cfg.TokenWeight * float64(d.Tokens)
		v += (st.w.Power + st.w.Toughness) * float64(d.Counters)
		v += cfg.DamageToOpponent * float64(d.DamageEachOpponent*len(st.opps))
	}
	return v
}

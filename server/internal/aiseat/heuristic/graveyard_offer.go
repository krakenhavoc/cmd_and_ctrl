package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// graveyard_offer.go — ADR 0187 §7 (owner decision 3, #3571): the bot as
// an offeree of an any-player offer paid by exiling cards from its own
// graveyard (Carrion Rats, Carrion Wurm: "any player may exile a card
// from their graveyard. If a player does, this creature assigns no combat
// damage this turn").
//
// The cards are a real price (a graveyard is flashback, escape and
// reanimation fuel, priced by fuelValue like dredge and devour price
// theirs), so the bot pays only when the creature is fighting IT. Paying
// to blunt an attacker aimed at someone else, or a blocker in a combat the
// bot is not in, spends fuel to help another seat. Its own creature's
// offer is declined by the general rule in choices.go.

const payActionExileGraveyard = "exile_graveyard"

// exilesFromMyGraveyard reports whether the prompt is an any-player offer
// priced in cards from the chooser's own graveyard.
func exilesFromMyGraveyard(ch *protocol.PendingChoiceView) bool {
	return ch != nil && ch.AnyPlayerOffer != nil && ch.PayCards != nil && ch.PayCards.Action == payActionExileGraveyard
}

// inCombatWithMe reports whether the battlefield creature `id` is
// attacking this seat, a planeswalker or battle it protects, or is
// blocking or blocked by a creature this seat controls.
func (st *state) inCombatWithMe(id string) bool {
	src := st.bf[id]
	if src == nil {
		return false
	}
	if src.AttackingTarget != "" {
		if src.AttackingTarget == st.me {
			return true
		}
		if t := st.bf[src.AttackingTarget]; t != nil && (t.Controller == st.me || t.ProtectorPlayer == st.me) {
			return true
		}
	}
	blocked := src.BlockingTargets
	if len(blocked) == 0 && src.BlockingTarget != "" {
		blocked = []string{src.BlockingTarget}
	}
	for _, a := range blocked {
		if atk := st.bf[a]; atk != nil && atk.Controller == st.me {
			return true
		}
	}
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller != st.me {
			continue
		}
		if c.BlockingTarget == id {
			return true
		}
		for _, a := range c.BlockingTargets {
			if a == id {
				return true
			}
		}
	}
	return false
}

package game

import "github.com/google/uuid"

// exiled_by_ability.go — "if it was exiled by an ability you controlled"
// (Evelyn, the Covetous, #5876).
//
// The zone-change event already says which player's resolving item moved
// a card (Event.CauseController, move_cause.go), but an event is gone by
// the time a cast permission asks about the card turns later. So the
// answer is written onto the card as it lands in exile
// (Card.ExiledByAbilityOf), the way the "exiled with" link is
// (stampExiledWithLocked), and MoveCard clears it with the other per-
// object exile marks (CR 400.7).
//
// Only an ABILITY's exile stamps it. The cause names the stack item, and
// an item is a spell or an ability (StackItem.Kind); the item is found
// in the resolving slot or still on the stack. A cause whose item is in
// neither place cannot be told from a spell's and stamps nothing, which
// is weaker than printed and never stronger.

// moveCause is the cause the event's route carries, the zero cause when
// the event has no route (a replacement-only exit).
func (ev *ReplacementEvent) moveCause() MoveCause {
	if ev == nil || ev.zoneRoute == nil {
		return MoveCause{}
	}
	return ev.zoneRoute.Cause
}

// abilityCauseControllerLocked is the controller of the ABILITY a move's
// cause names, uuid.Nil when the cause is not a resolving ability's
// effect. Caller must hold g.mu.
func (g *Game) abilityCauseControllerLocked(c MoveCause) uuid.UUID {
	if c.Kind != MoveCauseEffect || c.Item == uuid.Nil || c.Controller == uuid.Nil {
		return uuid.Nil
	}
	var item *StackItem
	if g.resolving != nil && g.resolving.item != nil && g.resolving.item.ID == c.Item {
		item = g.resolving.item
	} else {
		item = g.StackMeta[c.Item]
	}
	if item == nil || item.Kind == StackItemSpell || item.Kind == "" {
		return uuid.Nil
	}
	return c.Controller
}

// stampExiledByAbilityLocked writes Card.ExiledByAbilityOf onto the card
// that has just landed in `dst`, when `dst` is exile and the move was a
// resolving ability's. A no-op otherwise, and for a card that is not in
// `dst` (a move a replacement sent somewhere else). Called after
// MoveCard, which has already cleared any older mark. Caller must hold
// g.mu.
func (g *Game) stampExiledByAbilityLocked(dst *Zone, cardID uuid.UUID, c MoveCause) {
	if dst == nil || dst.Kind != ZoneExile {
		return
	}
	who := g.abilityCauseControllerLocked(c)
	if who == uuid.Nil {
		return
	}
	for i := range dst.Cards {
		if dst.Cards[i].InstanceID == cardID {
			dst.Cards[i].ExiledByAbilityOf = who
			return
		}
	}
}

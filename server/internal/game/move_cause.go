package game

import "github.com/google/uuid"

// move_cause.go — WHAT moved a card, on the zone-change event (#1320,
// ADR 0013 §5ad).
//
// Before this, EventZoneMove said where a card went and, for some
// routes, which player's action it was (Actor) and which card asked
// (Source). Neither answers the question a "whenever a spell or
// ability YOU CONTROL exiles one or more permanents" trigger asks
// (Ranar the Ever-Watchful): Actor is empty on the plain exile route,
// and Source is empty too, so an airbend, a Swords to Plowshares and a
// sandbox drag into exile all looked the same.
//
// # Where the cause comes from
//
// A card leaves a zone for one of five reasons, and the route names
// the reason when its caller knows it:
//
//   - a COST (CR 601.2h, 602.2b): exile_cost.go, return_cost.go and a
//     cost discard say so, and name the payer;
//   - a SPECIAL ACTION (CR 116.2): foretell and suspend;
//   - a RULE with no spell or ability behind it: a spell leaving the
//     stack as the last step of its own resolution (CR 608.2n), the
//     cleanup step's discard (CR 514.1);
//   - a MANUAL sandbox move, which is nobody's spell or ability;
//   - an EFFECT — a resolving spell or ability. This is the one the
//     route does NOT have to name: every effect exit in the engine
//     runs while that item is resolving, so routeCardToZoneLocked
//     fills it from Game.resolving (resolving_item.go). That slot is
//     set for spells AND abilities, and it survives every paused
//     continuation of the resolution that opened it, because a
//     resolution-time prompt blocks the table.
//
// Inferring the effect case rather than asking ~40 exit callers to
// thread the item through is the same choice #529 made for the
// CR 903.9 window: the defect would be structural if the cause were
// opened by individual movers, because a new mover would forget it.
// The explicit kinds exist precisely so the inference is never asked
// in the one window where Game.resolving is stale — after a
// resolution has finished and before play moves on, a player can pay
// a cost or perform a special action, and neither of those is the
// previous spell's doing.
//
// # What it does not cover (declared)
//
//   - The destroy / sacrifice / SBA exit (executeBattlefieldLeaveLocked)
//     carries no cause. A destruction a replacement turns into an exile
//     (Rest in Peace) is therefore not recorded as "the spell exiled
//     it". Weaker than printed for a Ranar reading it, never stronger.
//   - A zone move that bypasses the route (the leaving-player cleanup's
//     raw MoveCard) emits no zone-change event at all, so there is
//     nothing to stamp.

// MoveCauseKind is why a card changed zones. The zero value on an
// EVENT means "nothing recorded"; on a ROUTE it means "infer it"
// (see routeCardToZoneLocked).
type MoveCauseKind string

const (
	// MoveCauseEffect is a resolving spell or ability's instruction.
	// Event.CauseController is that item's controller and
	// Event.CauseItem its stack-item ID.
	MoveCauseEffect MoveCauseKind = "effect"

	// MoveCauseCost is a cost being paid for a spell or ability.
	// CauseController is the paying player.
	MoveCauseCost MoveCauseKind = "cost"

	// MoveCauseSpecialAction is a special action (CR 116.2) — foretell,
	// suspend. CauseController is the player taking it.
	MoveCauseSpecialAction MoveCauseKind = "special_action"

	// MoveCauseRule is a move the rules make with no spell or ability
	// behind it: a spell put into a graveyard as it finishes resolving,
	// the cleanup step's discard.
	MoveCauseRule MoveCauseKind = "rule"

	// MoveCauseManual is a sandbox move made by hand.
	MoveCauseManual MoveCauseKind = "manual"
)

// MoveCause is the cause a route carries onto its events.
type MoveCause struct {
	Kind MoveCauseKind

	// Controller is the player responsible: the resolving item's
	// controller for an effect, the payer for a cost, the player
	// taking a special action.
	Controller uuid.UUID

	// Item is the stack item whose resolution moved the card
	// (MoveCauseEffect only). Equal to the spell's card ID for a
	// spell, a minted ID for an ability — see StackItem.ID.
	Item uuid.UUID
}

// resolutionCauseLocked is the cause of a move made while a stack item
// is resolving, and the zero MoveCause when nothing is. Caller must
// hold g.mu.
func (g *Game) resolutionCauseLocked() MoveCause {
	if g.resolving == nil || g.resolving.item == nil {
		return MoveCause{}
	}
	it := g.resolving.item
	return MoveCause{Kind: MoveCauseEffect, Controller: it.Controller, Item: it.ID}
}

// stampCause copies a cause onto an event. A zero cause leaves the
// event exactly as it was.
func (c MoveCause) stampCause(ev *Event) {
	if c.Kind == "" {
		return
	}
	ev.Cause = c.Kind
	ev.CauseController = c.Controller
	ev.CauseItem = c.Item
}

// discardMoveCause maps a discard's CR 701.8a cause onto the move
// cause its events carry: a cost discard names its payer, the cleanup
// step's is a rule, and an effect's is left for routeCardToZoneLocked
// to fill from the resolving item.
func discardMoveCause(c DiscardCause, player uuid.UUID) MoveCause {
	switch c {
	case DiscardCauseCost:
		return MoveCause{Kind: MoveCauseCost, Controller: player}
	case DiscardCauseCleanup:
		return MoveCause{Kind: MoveCauseRule}
	}
	return MoveCause{}
}

// ExiledBySpellOrAbilityOf reports whether a zone-change event is a
// card moved INTO EXILE by a resolving spell or ability that `player`
// controls — the "a spell or ability you control exiles" clause
// (Ranar the Ever-Watchful, #1320). The source zone is the caller's
// to check: Ranar asks about permanents, so it also wants
// ev.OldZone == ZoneBattlefield.
func ExiledBySpellOrAbilityOf(ev Event, player uuid.UUID) bool {
	return ev.Kind == EventZoneMove && ev.NewZone == ZoneExile &&
		ev.Cause == MoveCauseEffect && player != uuid.Nil &&
		ev.CauseController == player
}

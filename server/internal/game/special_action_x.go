package game

import (
	"strconv"

	"github.com/google/uuid"
)

// special_action_x.go — #5249: a special action whose cost has an {X}.
//
// CR 107.3d: "If a cost associated with a special action, such as a
// suspend cost or a morph cost, has an {X} or an X in it, the value of X
// is chosen by the player taking the special action immediately before
// they pay that cost." Aurelia's Vindicator's "Disguise {X}{3}{W}" is
// the card that needs it, and CR 702.168e (CR 702.37f for morph) lets
// the permanent's other abilities read the X chosen.
//
// The shape is an activation's {X}, end to end, rather than a new one:
//
//   - the payload carries it (SpecialActionParams.XValue, wire x_value),
//     bounded by MaxAnnouncedX and refused for a cost with no {X};
//   - the payment multiplies it into the generic demand through the same
//     helper an activation pays with (payAbilityManaCostLocked), so a
//     strict payment refuses an X the seat cannot pay;
//   - the legal-move enumerator offers the largest affordable X and marks
//     the move's X open from 0 to that (ADR 0122 §6.2), and the view
//     stamps the same ceiling on the row (SpecialActionXCeilingForEffect)
//     for the client's X picker;
//   - the performer gets it on SpecialAction.X, and turn_face_up puts it
//     on EventTurnedFaceUp's Amount, as cycling puts its X on EventCycle's
//     (ADR 0062's 2026-10-13 amendment).

// SpecialActionDemandsX reports whether taking this offer means choosing
// an X: it is a turn_face_up offer whose cost has an {X} in it. An
// unreadable cost demands nothing (the payment refuses it anyway).
//
// turn_face_up only, deliberately. It is the kind #5249 built the X for:
// its performer hands the X on (EventTurnedFaceUp's Amount). Suspend's
// "Suspend X—{X}…" also puts X time counters on the card, which its
// performer does not read, so an {X} there stays at 0 rather than being
// charged for counters that never arrive.
func SpecialActionDemandsX(sa SpecialAction) bool {
	if sa.Kind != SpecialActionTurnFaceUp || sa.Cost == "" {
		return false
	}
	parsed, err := ParseCost(sa.Cost)
	return err == nil && parsed.XSlots > 0
}

// specialActionLogLabel is the log line's label for an action as TAKEN:
// the offer's own label, plus the X chosen when its cost has one, so the
// table hears the price that was paid ("Turn face up {X}{3}{W} for X=2").
// CR 702.168d has the player show everyone the disguise cost; the X is
// part of what they pay.
func specialActionLogLabel(sa SpecialAction) string {
	if !SpecialActionDemandsX(sa) {
		return sa.Label
	}
	return sa.Label + " for X=" + strconv.Itoa(sa.X)
}

// SpecialActionXCeilingForEffect is the largest X `actor` can pay for an
// {X} in this offer's cost right now, from the floating pool and an
// auto-tap plan together — the question the engine's strict auto-tapped
// payment asks — priced through the same CR 601.2f pass and spend context
// the engine charges with. ok is false when the offer demands no X, or
// when not even X=0 can be paid.
//
// The view stamps it on the row so the client's X picker can hold X
// under it. The scan stops at the first X that cannot be paid: every
// extra point of X is one more generic mana, so the cost only grows.
//
// Caller must hold g.mu (read or write). Read-only.
func (g *Game) SpecialActionXCeilingForEffect(actor uuid.UUID, card Card, sa SpecialAction) (int, bool) {
	if !SpecialActionDemandsX(sa) {
		return 0, false
	}
	p := g.playerByIDLocked(actor)
	if p == nil {
		return 0, false
	}
	cost, err := g.SpecialActionManaCostForEffect(actor, card, sa.Kind, sa)
	if err != nil {
		return 0, false
	}
	spend := specialActionSpendContext(sa.Kind)
	payable := func(x int) bool {
		paid := g.CostAsPaidByForEffect(actor, spend, cost, x)
		return p.ManaPool.CanPayFor(paid, x, spend) || g.AutoTapTopUpForEffectExcluding(actor, paid, x, spend, nil)
	}
	best := -1
	for x := 0; x <= MaxAnnouncedX && payable(x); x++ {
		best = x
	}
	return best, best >= 0
}

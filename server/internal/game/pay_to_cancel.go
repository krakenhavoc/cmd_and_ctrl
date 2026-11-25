package game

import (
	"slices"

	"github.com/google/uuid"
)

// pay_to_cancel.go — "… unless they pay {N} before that step" (#6220):
// Nafs Asp's "that player loses 1 life at the beginning of their next draw
// step unless they pay {1} before that draw step", Sabertooth Cobra's "The
// player gets another poison counter at the beginning of their next upkeep
// unless they pay {2} before that step."
//
// The consequence is an ordinary delayed trigger (CR 603.7) at the
// beginning of that player's next step, and the payment rides it as data
// (DelayedTrigger.PayToCancel). The player may pay at any time they have
// priority while it waits (CR 117.1: the player with priority may take
// actions) — on any turn, with anything on the stack — through the
// `special_action` verb with kind pay_to_cancel and the trigger's ID as
// card_id. Paying removes the trigger, so nothing happens at that step.
// The window closes as the step begins: the trigger fires then and leaves
// the queue, and a payment refers to nothing. Paying is not casting or
// activating, so split second does not stop it and the mana spent is
// unrestricted-only (the zero spend context, as for a special action).

// SpecialActionPayToCancel pays a waiting delayed trigger's PayToCancel
// cost and removes the trigger (#6220). Its card_id is the trigger's ID.
const SpecialActionPayToCancel SpecialActionKind = "pay_to_cancel"

// PayToCancel is the payment on a delayed trigger: Player may pay Cost
// (brace notation, "{1}") before the trigger fires, and then it never
// does.
type PayToCancel struct {
	Player uuid.UUID `json:"player"`
	Cost   string    `json:"cost"`
}

func clonePayToCancel(p *PayToCancel) *PayToCancel {
	if p == nil {
		return nil
	}
	out := *p
	return &out
}

// PayToCancelOffer is one payment a player may make now: the waiting
// trigger's ID, the cost, what the trigger would do, and the card it is
// from.
type PayToCancelOffer struct {
	ID         uuid.UUID
	Player     uuid.UUID
	Cost       string
	Label      string
	SourceName string
	At         Step
}

// PayToCancelOffersLocked lists the payments waiting on `player`, in queue
// order. Nil when there are none, which is nearly always. Caller must hold
// g.mu.
func (g *Game) PayToCancelOffersLocked(player uuid.UUID) []PayToCancelOffer {
	var out []PayToCancelOffer
	for _, dt := range g.DelayedTriggers {
		if dt == nil || dt.PayToCancel == nil || dt.PayToCancel.Player != player {
			continue
		}
		name := ""
		if c, _ := g.findCardAndZoneLocked(dt.SourceCardID); c != nil {
			name = c.Name
		}
		out = append(out, PayToCancelOffer{
			ID:         dt.ID,
			Player:     player,
			Cost:       dt.PayToCancel.Cost,
			Label:      dt.Label,
			SourceName: name,
			At:         dt.At,
		})
	}
	return out
}

// PayToCancelCostForEffect is the parsed cost of the waiting payment `id`
// for `player`, for the enumerator's affordability check. False when there
// is no such payment. Caller must hold g.mu.
func (g *Game) PayToCancelCostForEffect(player, id uuid.UUID) (ParsedCost, bool) {
	dt := g.payToCancelTriggerLocked(player, id)
	if dt == nil {
		return ParsedCost{}, false
	}
	cost, err := ParseCost(dt.PayToCancel.Cost)
	if err != nil {
		return ParsedCost{}, false
	}
	return cost, true
}

func (g *Game) payToCancelTriggerLocked(player, id uuid.UUID) *DelayedTrigger {
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.ID == id && dt.PayToCancel != nil && dt.PayToCancel.Player == player {
			return dt
		}
	}
	return nil
}

// performPayToCancelLocked is the pay_to_cancel special action: `p` pays
// the waiting trigger `id`'s cost and the trigger is removed. Priority is
// the dispatcher's check (requirePriorityHolder). Caller must hold g.mu.
func (g *Game) performPayToCancelLocked(p *Player, id uuid.UUID, params SpecialActionParams) error {
	if params.SacrificeID != uuid.Nil || params.Door != DoorNone || params.Cost != "" || params.Route != "" || params.XValue != 0 {
		return ErrInvalidParam
	}
	dt := g.payToCancelTriggerLocked(p.ID, id)
	if dt == nil {
		return ErrSpecialActionNotOffered
	}
	cost, err := ParseCost(dt.PayToCancel.Cost)
	if err != nil {
		return ErrInvalidParam
	}
	g.beginEventBatchLocked()
	name := "the payment"
	if c, _ := g.findCardAndZoneLocked(dt.SourceCardID); c != nil {
		name = c.Name
	}
	if _, err := g.payAbilityManaCostLocked(p, dt.SourceCardID, name, cost, ActivateAbilityParams{
		Strict:  params.Strict,
		AutoTap: params.AutoTap,
	}, ManaSpendContext{}, nil); err != nil {
		return err
	}
	g.DelayedTriggers = slices.DeleteFunc(slices.Clone(g.DelayedTriggers), func(d *DelayedTrigger) bool { return d != nil && d.ID == id })
	g.EmitEvent(Event{
		Kind:   EventSpecialAction,
		Actor:  p.ID,
		CardID: dt.SourceCardID,
		Label:  "Pay " + dt.PayToCancel.Cost + ": " + payToCancelWhat(dt),
	})
	// CR 117.3c: the player keeps priority, and passes made before the
	// action no longer count.
	g.noteActionTakenLocked(p.ID)
	g.runStateChecksLocked()
	return nil
}

func payToCancelWhat(dt *DelayedTrigger) string {
	if dt.Label != "" {
		return dt.Label
	}
	return "the delayed trigger"
}

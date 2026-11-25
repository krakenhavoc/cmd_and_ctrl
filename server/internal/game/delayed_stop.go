package game

import (
	"slices"

	"github.com/google/uuid"
)

// delayed_stop.go — CR 116.2c: "Some effects allow a player to take an
// action at a later time, usually to end a continuous effect or to stop a
// delayed triggered ability from triggering. Doing so is a special
// action. A player can take such an action any time they have priority,
// unless that effect specifies another timing restriction, for as long as
// the effect allows it." (#6247)
//
// Quenchable Fire: "It deals an additional 3 damage to that player or
// planeswalker at the beginning of your next upkeep step unless that
// player or that planeswalker's controller pays {U} before that step."
//
// The delayed trigger carries the offer as data (DelayedTrigger.Stop):
// who may pay, what, and the row they read. The special action
// stop_delayed_trigger names the trigger by its ID in card_id, pays the
// mana as every special action's mana is paid, and removes the trigger
// from the queue, so it never triggers. It is offered "before that step"
// for exactly as long as the trigger is waiting: once it has triggered
// it has left the queue and the offer goes with it.

// SpecialActionStopDelayedTrigger is CR 116.2c's special action: pay to
// stop a delayed triggered ability from triggering.
const SpecialActionStopDelayedTrigger SpecialActionKind = "stop_delayed_trigger"

// DelayedTriggerStop is the CR 116.2c escape a delayed trigger offers:
// Player may pay Cost, any time they have priority before it triggers, to
// stop it. Plain data, carried by the snapshot with the trigger.
type DelayedTriggerStop struct {
	// Player is who may pay ("that player or that planeswalker's
	// controller", fixed as the spell resolves).
	Player uuid.UUID `json:"player"`
	// Cost is the mana, in brace notation ("{U}").
	Cost string `json:"cost"`
	// Label is the row as the player reads it ("Pay {U}: Quenchable
	// Fire deals no additional damage to you").
	Label string `json:"label,omitempty"`
}

// StopDelayedTriggerOffer is one delayed trigger a player may pay to stop
// right now.
type StopDelayedTriggerOffer struct {
	// ID is the delayed trigger's; the action's card_id.
	ID uuid.UUID
	// SourceName is the name of the card that made it.
	SourceName string
	// Label is DelayedTriggerStop.Label, or the trigger's own label.
	Label string
	// Cost is the mana, in brace notation.
	Cost string
}

// StopDelayedTriggerOffersLocked is every CR 116.2c stop `player` may
// take right now, in queue order. THE accessor: the engine's check, the
// legal-move enumerator and the view all ask here. Nil for a player who
// has left the game and for nearly every game.
//
// Caller must hold g.mu (read or write).
func (g *Game) StopDelayedTriggerOffersLocked(player uuid.UUID) []StopDelayedTriggerOffer {
	if p := g.playerByIDLocked(player); p == nil || p.Eliminated {
		return nil
	}
	var out []StopDelayedTriggerOffer
	for _, dt := range g.DelayedTriggers {
		if dt == nil || dt.Stop == nil || dt.Stop.Player != player {
			continue
		}
		label := dt.Stop.Label
		if label == "" {
			label = "Pay " + dt.Stop.Cost + ": stop " + dt.Label
		}
		out = append(out, StopDelayedTriggerOffer{
			ID:         dt.ID,
			SourceName: g.delayedSourceNameLocked(dt),
			Label:      label,
			Cost:       dt.Stop.Cost,
		})
	}
	return out
}

// delayedSourceNameLocked is the name of the card a delayed trigger came
// from, wherever that card is now, or the trigger's label.
func (g *Game) delayedSourceNameLocked(dt *DelayedTrigger) string {
	if c, ok := g.LookupCardForEffect(dt.SourceCardID); ok && c.Name != "" {
		return c.Name
	}
	return dt.Label
}

// performStopDelayedTriggerLocked is the stop_delayed_trigger performer:
// find the offer, pay its mana, take the trigger out of the queue. A
// refusal leaves everything as it was; the action uses no stack and the
// player keeps priority (CR 116.3).
//
// Caller must hold g.mu (write) and have run the dispatcher's priority
// check.
func (g *Game) performStopDelayedTriggerLocked(p *Player, id uuid.UUID, params SpecialActionParams) error {
	if params.Cost != "" || params.Door != DoorNone || params.Route != "" || params.SacrificeID != uuid.Nil {
		return ErrInvalidParam
	}
	offers := g.StopDelayedTriggerOffersLocked(p.ID)
	at := slices.IndexFunc(offers, func(o StopDelayedTriggerOffer) bool { return o.ID == id })
	if at < 0 {
		return ErrSpecialActionNotOffered
	}
	offer := offers[at]
	cost, err := ParseCost(offer.Cost)
	if err != nil {
		return ErrInvalidParam
	}
	g.beginEventBatchLocked()
	if _, err := g.payAbilityManaCostLocked(p, uuid.Nil, offer.SourceName, cost, ActivateAbilityParams{
		Strict:  params.Strict,
		AutoTap: params.AutoTap,
	}, specialActionSpendContext(SpecialActionStopDelayedTrigger), nil); err != nil {
		return err
	}
	g.DelayedTriggers = slices.DeleteFunc(g.DelayedTriggers, func(dt *DelayedTrigger) bool {
		return dt != nil && dt.ID == id
	})
	g.EmitEvent(Event{Kind: EventSpecialAction, Actor: p.ID, Label: offer.Label})
	g.noteActionTakenLocked(p.ID)
	g.runStateChecksLocked()
	return nil
}

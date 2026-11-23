package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tiana, Ship's Caretaker — Legendary Creature — Angel Artificer
// {3}{R}{W}, 3/3 (Dominaria Remastered, #5948):
//
//	"Flying, first strike
//	 Whenever an Aura or Equipment you control is put into a graveyard from
//	 the battlefield, you may return that card to its owner's hand at the
//	 beginning of the next end step."
//
// The trigger watches every permanent that leaves for a graveyard, and
// keeps the ones that were an Aura or an Equipment under Tiana's
// controller's control as they last existed (CR 603.10a: the event's
// last-known controller). The "may" is asked as the ability resolves,
// before the delayed trigger is scheduled; the delayed trigger then returns
// that card from the graveyard to its owner's hand at the next end step,
// and does nothing if the card has gone since.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "80f8fb4c-cc6d-4b41-abb3-471dc96d4b2a",
		Name:            "Tiana, Ship's Caretaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.NewZone != game.ZoneGraveyard || ev.CardID == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || effects.IsToken(c) {
					return false
				}
				return (c.HasSubtype("Aura") || c.HasSubtype("Equipment")) && effects.LeftUnderControlOf(ev, c) == source.Controller
			},
			Key: "Tiana, Ship's Caretaker — you may return that card to its owner's hand at the beginning of the next end step",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.MayChoice{
					Question: "Tiana, Ship's Caretaker — return that card to its owner's hand at the beginning of the next end step?",
					Then:     tianaReturn,
					Carry:    []uuid.UUID{item.Trigger.Event.CardID},
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

var tianaReturn = effects.MayYes("may/tiana-ships-caretaker-return", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		Label: "Tiana, Ship's Caretaker — return the card to its owner's hand",
		Cards: []uuid.UUID{r.Carry[0]},
		Body:  effects.ReturnListedGraveyardToHandBody,
	}.Apply(ctx)
})

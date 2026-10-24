package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// First Day of Class — Instant {1}{R} (EDHREC rank 9897):
//
//	"Whenever a creature you control enters this turn, put a +1/+1
//	 counter on it and it gains haste until end of turn.
//	 Learn. (You may reveal a Lesson card you own from outside the game
//	 and put it into your hand, or discard a card to draw a card.)"
//
// The first sentence creates a delayed triggered ability (CR 603.7a)
// that watches creatures ENTERING under your control until the turn
// ends (CR 603.7b), and triggers for each one (CR 603.2c), carrying the
// creature as its payload (delayed_creature_events.go, the Shriveling
// Rot shape). The counter and the haste go to that creature only if it
// is still the same object on the battlefield when the trigger
// resolves (CR 603.7c). A creature that entered before First Day of
// Class resolved is not affected: the ability did not exist yet.
//
// The delayed trigger is set up before learn, as printed (CR 701.48a,
// learn.go). Its duration is the turn (CR 603.7b), so it ends with it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0b5c772-e029-4b8d-be67-b1584700e377",
		Name:         "First Day of Class",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return DelayedOnEvent{
					Label:     "First Day of Class — put a +1/+1 counter on it and it gains haste until end of turn",
					On:        []game.EventKind{game.EventETB},
					Condition: eachCreatureYouControlEntersCondition,
					Body:      counterAndHasteForThePayloadBody,
					Repeats:   true,
				}.Apply(ctx)
			}),
			Learn{},
		),
	})
}

// eachCreatureYouControlEntered is the creature an EventETB names, while
// it is a creature the delayed trigger's controller controls.
func eachCreatureYouControlEntered(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) []uuid.UUID {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsCreature() || c.Controller != dt.Controller {
		return nil
	}
	return []uuid.UUID{ev.CardID}
}

// counterAndHasteForThePayload is "put a +1/+1 counter on it and it gains
// haste until end of turn": the payload's permanent, if it is still that
// object on the battlefield.
func counterAndHasteForThePayload(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	obj, ok := payloadObject(ctx)
	if !ok {
		return nil
	}
	c, ok := battlefieldCardForEffect(g, obj.ID)
	if !ok || c.ObjectEpoch != obj.Epoch {
		return nil
	}
	if err := (AddCounter{Target: obj.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{
		Target:   obj.ID,
		Keywords: []string{"haste"},
		Label:    "First Day of Class — haste until end of turn",
	}.Apply(ctx)
}

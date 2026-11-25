package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tribute.go — tribute (CR 702.104, #6251). The engine half is
// game/tribute.go. A card file writes:
//
//	Replacements: Tribute("Ornitharch", 2),
//	Triggered: []game.TriggeredAbility{
//	    WhenEntersIfTributeWasntPaid("Ornitharch — create two 1/1 white Bird tokens with flying", …),
//	},
//
// CR 702.104a: "As this creature enters, choose an opponent. That
// player may put an additional N +1/+1 counters on it as it enters."
// CR 702.104b: "if tribute wasn't paid" is true when that opponent didn't
// have the creature enter with the counters.

// Tribute is "Tribute N": the controller's choice of opponent, then that
// opponent's "may" (both before the creature enters, CR 614.12a). `name`
// is the card's name, for the prompts and the CR 616 labels.
//
// The two effects apply one after the other: the payment is not
// applicable until an opponent has been chosen. With one opponent the
// choice is made without asking; an entry that cannot pause takes the
// first opponent after the controller and does not ask the payment, so
// the tribute is not paid.
func Tribute(name string, n int) []game.ReplacementEffect {
	ownEntry := func(ev *game.ReplacementEvent, src *game.Card) bool {
		return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
			src != nil && ev.CardID == src.InstanceID
	}
	counters := "a +1/+1 counter"
	if n != 1 {
		counters = fmt.Sprintf("%d +1/+1 counters", n)
	}
	return []game.ReplacementEffect{
		{
			Watches:         []game.EventKind{game.EventZoneMove},
			SelfReplacement: true,
			EntryOnly:       true,
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ownEntry(ev, src)
			},
			Controller: EnteringPermanentChooser,
			EntryController: &game.EntryControllerChoice{
				Purpose:  game.ControlForTribute,
				Question: fmt.Sprintf("Tribute %d — choose an opponent to decide whether to put %s on %s", n, counters, name),
			},
			Label:    fmt.Sprintf("%s: tribute %d, choose an opponent", name, n),
			RowLabel: fmt.Sprintf("Tribute %d.", n),
		},
		{
			Watches:         []game.EventKind{game.EventZoneMove},
			SelfReplacement: true,
			EntryOnly:       true,
			Optional:        true,
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ownEntry(ev, src) && ev.TributeOpponent != uuid.Nil
			},
			Controller: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID {
				return ev.TributeOpponent
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.AddCounterAtETB(game.CounterPlusOne, n)
				ev.EntersTributePaid = true
				return nil
			},
			PromptQuestion: fmt.Sprintf("Tribute %d — put %s on %s as it enters? If you don't, its \"tribute wasn't paid\" ability happens.", n, counters, name),
			Label:          fmt.Sprintf("%s: tribute %d, pay it", name, n),
			RowLabel:       fmt.Sprintf("Tribute %d.", n),
		},
	}
}

// TributeWasntPaid is the trigger-time half of "if tribute wasn't paid"
// (CR 702.104b, CR 603.4), read off the creature itself.
func TributeWasntPaid(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && !source.TributePaid
}

// CondTributeWasntPaid is the resolution half (CR 603.4): the creature's
// record, or its last-known information once it has left (CR 608.2h).
type CondTributeWasntPaid struct{}

func (CondTributeWasntPaid) Holds(ctx *Context) bool {
	info, ok := ctx.SourcePermanent()
	return ok && !info.TributePaid
}

func (CondTributeWasntPaid) Condition() (game.FootprintStep, bool) {
	return CondStep(game.FootprintCondSource, 0), true
}

// WhenEntersIfTributeWasntPaid is "When this creature enters, if
// tribute wasn't paid, <effect>": one ETB trigger whose intervening "if"
// is checked as it triggers and again as it resolves (CR 603.4).
func WhenEntersIfTributeWasntPaid(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondTributeWasntPaid{}, On(game.EventETB, AllOf(Self, TributeWasntPaid), label, effect))
}

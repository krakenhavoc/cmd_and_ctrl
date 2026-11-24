package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reyhan, Last of the Abzan — Legendary Creature — Human Warrior {1}{B}{G},
// 0/0 (Commander Legends, #5507):
//
//	"Reyhan enters with three +1/+1 counters on it.
//	 Whenever a creature you control dies or is put into the command zone, if
//	 it had one or more +1/+1 counters on it, you may put that many +1/+1
//	 counters on target creature.
//	 Partner"
//
// The first line is Mossborn Hydra's entry counters. The trigger watches a
// creature you controlled leave the battlefield for a graveyard or for the
// command zone (CR 903.9), Reyhan herself included; "it had one or more
// +1/+1 counters" is read from the departure record (last-known information,
// CR 603.10a) when it triggers (the intervening "if", CR 603.4: that record
// cannot change) and again for "that many" as it resolves. The "you may" is the
// controller's, and the target is chosen as the trigger goes on the stack.
// Partner is a deck-construction rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e7df01c-75af-440d-97fd-1b1b794281b9",
		Name:         "Reyhan, Last of the Abzan",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.AddCounterAtETB(game.CounterPlusOne, 3)
				return nil
			},
			Label: "Reyhan, Last of the Abzan: enters with three +1/+1 counters",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventLTB || (ev.NewZone != game.ZoneGraveyard && ev.NewZone != game.ZoneCommand) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || !effects.LeftAsType(ev, c, "creature") || effects.LeftUnderControlOf(ev, c) != source.Controller {
					return false
				}
				return g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
			},
			Targets: effects.TargetCreature("target creature"),
			Key:     "Reyhan, Last of the Abzan — put that many +1/+1 counters on target creature",
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Reyhan, Last of the Abzan — put that many +1/+1 counters on target creature?",
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				left, ok := ctx.TriggeringPermanent()
				if !ok {
					return nil
				}
				n := left.Counters[game.CounterPlusOne]
				id := effects.FirstLegalCardTarget(ctx)
				if n <= 0 || id == uuid.Nil {
					return nil
				}
				return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
			},
		}},
	})
}

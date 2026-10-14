package effects

import (
	"errors"
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Luminous Broodmoth — Creature — Insect {2}{W}{W}, 3/4:
//
//	"Flying
//	 Whenever a creature you control without flying dies, return it to
//	 the battlefield under its owner's control with a flying counter on
//	 it."
//
// "Without flying" and "you control" are read as the creature last
// existed on the battlefield (CR 603.10a): the controller off the exit
// event, the abilities off its departure record
// (Game.LastKnownPermanentForEffect), so a creature that had flying
// only from an anthem or a counter does not come back, and one that
// lost flying to an effect does. The flying counter rides the entry
// (ReturnFromGraveyardWithCountersForEffect), so the creature comes
// back with flying and the next death does not trigger again. A token
// is gone by the time the trigger resolves, and a card that left the
// graveyard in response stays where it is (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "28c7c816-07e7-42fb-923c-bf149ba28b38",
		Name:            "Luminous Broodmoth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, aCreatureYouControlWithoutFlyingDied,
				"Luminous Broodmoth — return it to the battlefield with a flying counter",
				func(g *game.Game, item *game.StackItem) error {
					_, err := g.ReturnFromGraveyardWithCountersForEffect(item.Trigger.Event.CardID, uuid.Nil, false,
						map[string]int{game.CounterFlying: 1})
					if errors.Is(err, game.ErrCardNotFound) {
						return nil
					}
					return err
				}),
		},
	})
}

// aCreatureYouControlWithoutFlyingDied is the Broodmoth's condition: a
// creature you controlled died, and it did not have flying as it last
// existed on the battlefield.
func aCreatureYouControlWithoutFlyingDied(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !ACreatureYouControlDied(ev, source, lki, g) {
		return false
	}
	if info, ok := g.LastKnownPermanentForEffect(ev.CardID); ok {
		return !slices.Contains(info.Characteristic.Abilities, "flying")
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !slices.Contains(c.Keywords, "flying")
}

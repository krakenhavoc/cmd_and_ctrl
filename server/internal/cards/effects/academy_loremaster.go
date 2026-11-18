package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Academy Loremaster — Creature — Human Wizard {U}{U}, 2/3 (Dominaria United):
//
//	"At the beginning of each player's draw step, that player may draw an
//	 additional card. If they do, spells they cast this turn cost {2} more to
//	 cast."
//
// Edric's political "may": the trigger is Loremaster's, on its controller's
// side of the stack, but the draw step's player (the event's Actor) is the one
// asked, and the one who draws. The "yes" is a choice to draw one card
// (MayDraw), so a player a draw limit stops isn't asked (CR 121.2b).
//
// Drawing gives that player a turn-long tax: a NextSpellPromise with
// EveryMatchingSpell and Increase 2, held on the drawing player and ending
// with the turn (an ordinary CR 601.2f increase on every spell they cast, any
// zone). It stands whether or not Loremaster stays on the battlefield, as
// "this turn" says.
//
// No simplification.
func init() {
	const label = "Academy Loremaster — that player may draw an additional card; if they do, their spells cost {2} more this turn"
	Register(Spec{
		OracleID:     "7d540860-ae94-4773-a526-aaffca3eb057",
		Name:         "Academy Loremaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{MayDraw(game.TriggeredAbility{
			Watches: []game.EventKind{game.EventBeginDrawStep},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBeginDrawStep && ev.Actor != uuid.Nil
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				if ev.Actor == uuid.Nil {
					return nil
				}
				return game.NewTriggeredItem(source, label)
			},
			Effect: academyLoremasterResolve,
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Academy Loremaster — draw an additional card? (Your spells cost {2} more this turn.)",
				Chooser: func(ev game.Event, _ *game.Card, _ *game.Game) uuid.UUID {
					return ev.Actor
				},
			},
		}, 1)},
	})
}

func academyLoremasterResolve(g *game.Game, item *game.StackItem) error {
	player := triggeringActor(item)
	if player == uuid.Nil {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: player, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return GrantNextSpellPromise{Player: player, From: "Academy Loremaster", Promise: game.NextSpellPromise{
		Increase:           2,
		EveryMatchingSpell: true,
		Text:               "Spells you cast this turn cost {2} more to cast.",
	}}.Apply(ctx)
}

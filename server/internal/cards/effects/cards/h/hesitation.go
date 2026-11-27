package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hesitation — Enchantment {1}{U} (Stronghold, S82 OLD b750, #5627):
//
//	"When a player casts a spell, sacrifice this enchantment and counter that
//	 spell."
//
// One trigger on every cast by any player (Standstill's), with two instructions
// joined by "and", not "if you do": the sacrifice, then the counter. Another
// Hesitation, or the same one answering a second cast, still counters its
// spell even when this one is already gone; a spell that has left the stack in
// response is simply not countered (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11eb31d9-cbcb-4ec5-abab-dfa646addbba",
		Name:         "Hesitation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventCast && ev.Actor != uuid.Nil
			}, "Hesitation — sacrifice this enchantment and counter that spell",
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.SacrificeThisIfStillOnBattlefield(g, item); err != nil {
						return err
					}
					return effects.CounterTarget{StackID: item.Trigger.Event.CardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mondronen Shaman // Tovolar's Magehunter — {3}{R} Creature — Human
// Shaman Werewolf 3/2 // Creature — Werewolf 5/5 (#4071, ADR 0132
// amendment 2026-10-16):
//
//	Front: "At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Whenever an opponent casts a spell, this creature deals 2
//	        damage to that player.
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// "That player" is the caster, read off the trigger's event as it
// resolves. The damage is the Magehunter's, dealt even if it has left
// the battlefield by then.
//
// No simplification.
func init() {
	const oracle = "f94cb722-7e46-4569-b474-9654ecf7ec14"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Mondronen Shaman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Mondronen Shaman"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Tovolar's Magehunter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(nil), "Tovolar's Magehunter — 2 damage to that player",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: item.Trigger.Event.Actor, Amount: 2}.Apply(effects.NewContext(g, item))
				}),
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Tovolar's Magehunter"),
		},
	})
}

package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodline Recollector // Ancestral Craving — Creature — Vampire
// Warlock {1}{B}, 2/2 // Instant {B} (preparation card, CR 722):
//
//	"At the beginning of each end step, if three or more creatures died
//	 this turn, this creature becomes prepared. (While it's prepared, you
//	 may cast a copy of its spell. Doing so unprepares it.)"
//
//	Ancestral Craving — "Target player draws three cards and loses 3
//	 life."
//
// "Each end step" is every player's, so the trigger does not read the
// event's actor. The intervening if (CR 603.4) is checked when the end
// step begins and again as the trigger resolves, off the table-wide
// per-turn tally of creatures that died. It does not enter prepared, so
// it is an ordinary 2/2 until the turn it has watched three die.
//
// No simplification.
func init() {
	const id = "c8e9a7e1-28ae-40e3-8715-1bb0aa2057c7"
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Bloodline Recollector",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventBeginEndStep && effects.B11CreaturesDiedThisTurn(g) >= 3
			}, "Bloodline Recollector — becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					if effects.B11CreaturesDiedThisTurn(g) < 3 {
						return nil
					}
					return effects.BecomePrepared{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Ancestral Craving",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 3, LifeLoss: 3}),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B32TargetPlayerDrawsAndLosesLife(item, ctx, 3, 3)
		},
	})
}

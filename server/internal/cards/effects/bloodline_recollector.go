package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     id,
		Name:         "Bloodline Recollector",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventBeginEndStep && b11CreaturesDiedThisTurn(g) >= 3
			}, "Bloodline Recollector — becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					if b11CreaturesDiedThisTurn(g) < 3 {
						return nil
					}
					return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Ancestral Craving",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		Purpose:      ForTargets(game.TargetPurpose{Slot: 0, Draws: 3, LifeLoss: 3}),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b32TargetPlayerDrawsAndLosesLife(item, ctx, 3, 3)
		},
	})
}

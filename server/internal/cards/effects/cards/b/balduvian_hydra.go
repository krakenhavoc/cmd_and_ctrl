package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Balduvian Hydra — Creature — Hydra {X}{R}{R}, 0/1 (Ice Age, S82 OLD batch 845, #5636):
//
//	"This creature enters with X +1/+0 counters on it.
//	 Remove a +1/+0 counter from this creature: Prevent the next 1 damage that
//	 would be dealt to it this turn.
//	 {R}{R}{R}: Put a +1/+0 counter on this creature. Activate only during
//	 your upkeep."
//
// The X counters are the CR 614.1c entry clause (XCounters); a "+1/+0"
// counter is read as that much power by the engine's counter reader. The
// shield is Wojek Apothecary's one-point PreventNextDamage (CR 615.7) pinned
// to the Hydra, its cost the counter itself paid at announce; the last
// counter leaves a 0/1. A Hydra that has left by resolution has nothing to
// shield.
//
// No simplification.
func init() {
	const kind = "+1/+0"
	effects.Register(effects.Spec{
		OracleID:                   "29a12f75-d66d-457f-85c4-6b9450adc0d8",
		Name:                       "Balduvian Hydra",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(kind)},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Remove a +1/+0 counter from this creature: Prevent the next 1 damage that would be dealt to it this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.RemoveCountersFromThis(kind, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
						return nil
					}
					return effects.PreventNextDamage{Target: item.SourceCardID, Amount: 1, Label: "Balduvian Hydra: prevent the next 1 damage"}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:     "{R}{R}{R}: Put a +1/+0 counter on this creature. Activate only during your upkeep.",
				Purpose:   game.Purpose{Answers: game.AnswerPump},
				Cost:      effects.ManaCost("{R}{R}{R}"),
				Condition: effects.DuringYourUpkeep(),
				Effect:    effects.PutACounterOnThis(kind),
			},
		},
	})
}

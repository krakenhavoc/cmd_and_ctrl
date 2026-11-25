package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Planebound Accomplice — Creature — Human Wizard {2}{R}, 1/3 (Modern
// Horizons):
//
//	"{R}: You may put a planeswalker card from your hand onto the battlefield.
//	 Sacrifice it at the beginning of the next end step."
//
// Meek Attack's put-from-hand and delayed sacrifice for a planeswalker card:
// the card is put, never cast (CR 701.18a), so counterspells and cast triggers
// see nothing, and it enters with its starting loyalty. "You may" is a real
// decline and a player with no planeswalker card is not prompted. The delayed
// trigger sacrifices that same object at the next end step only if its
// controller still controls it (CR 603.7c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01076e04-3106-4968-8ef4-b26fc3786699",
		Name:         "Planebound Accomplice",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: You may put a planeswalker card from your hand onto the battlefield. Sacrifice it at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PutFromHandOntoBattlefield{
					Match:    effects.Planeswalker(),
					Optional: true,
					Label:    "Planebound Accomplice — you may put a planeswalker card from your hand onto the battlefield",
					ThenTail: planeboundAccompliceEntered,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// planeboundAccompliceEntered is "sacrifice it at the beginning of the next end
// step" for the planeswalker that entered. An on-disk key.
var planeboundAccompliceEntered = effects.PutFromHandTail("planebound-accomplice/entered", func(ctx *effects.Context, res effects.PutFromHandResult) error {
	if res.Entered == uuid.Nil {
		return nil
	}
	c, ok := effects.BattlefieldCardForEffect(ctx.Game, res.Entered)
	if !ok {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		Label:  "Planebound Accomplice — sacrifice it",
		Body:   effects.StdB66MeekAttackSacrifice,
		Params: game.EffectParams{Object: game.ObjectRef{ID: res.Entered, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})

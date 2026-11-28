package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Daredevil Dragster — Artifact — Vehicle {3}, 4/4 (Aether Revolt, S82 older
// sets, #5536):
//
//	"At end of combat, if this Vehicle attacked or blocked this combat, put a
//	 velocity counter on it. Then if it has two or more velocity counters on
//	 it, sacrifice it and draw two cards.
//	 Crew 2"
//
// The intervening "if" is read as the end of combat step begins and again
// as the ability resolves (CR 603.4): the combat phase's records say the
// Vehicle was declared as an attacker or a blocker, and stay true for one
// that has since left combat. The counter and the "then if" are one
// sentence, so a second counter sacrifices it in the same resolution, and
// the draw follows the sacrifice only when the Vehicle was really
// sacrificed (a Vehicle already gone draws nothing).
//
// No simplification.
func init() {
	const label = "Daredevil Dragster — put a velocity counter on it; then if it has two or more, sacrifice it and draw two cards"
	effects.Register(effects.Spec{
		OracleID:     "6aa96a6e-9a88-4662-aa26-b4c16c0ef650",
		Name:         "Daredevil Dragster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "this Vehicle attacked or blocked this combat",
				Test: func(ctx *effects.Context) bool {
					ref, ok := ctx.SourceRef()
					return ok && daredevilDragsterWasInCombat(ctx.Game, ref)
				},
			}, effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepEndCombat, false),
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return daredevilDragsterWasInCombat(g, game.ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch})
				}),
				label, daredevilDragsterEndOfCombat)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Daredevil Dragster"),
		}},
	})
}

// daredevilDragsterWasInCombat is "this Vehicle attacked or blocked this
// combat" for the object `ref` names: declared as an attacker in the combat
// phase in progress, or as a blocker in it. Both are records of the phase,
// so they stay true for a Vehicle that has left combat.
//
// Caller holds g.mu.
func daredevilDragsterWasInCombat(g *game.Game, ref game.ObjectRef) bool {
	return effects.OldB819WasInCombat(g, ref)
}

// daredevilDragsterEndOfCombat puts the velocity counter, then, as the same
// sentence says, sacrifices the Vehicle and draws two cards if it has two or
// more.
func daredevilDragsterEndOfCombat(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	if err := (effects.AddCounter{Target: item.SourceCardID, Kind: "velocity", N: 1}).Apply(ctx); err != nil {
		return err
	}
	if effects.CounterCountOn(g, item.SourceCardID, "velocity") < 2 {
		return nil
	}
	if err := effects.SacrificeThisIfStillOnBattlefield(g, item); err != nil {
		return err
	}
	return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
}

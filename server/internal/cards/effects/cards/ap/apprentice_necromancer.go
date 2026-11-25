package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Apprentice Necromancer — Creature — Zombie Wizard {1}{B}, 1/1 (Commander
// 2017, #5533):
//
//	"{B}, {T}, Sacrifice this creature: Return target creature card from
//	 your graveyard to the battlefield. That creature gains haste. At the
//	 beginning of the next end step, sacrifice it."
//
// Whip of Erebos's return and haste, with the sacrifice of Sneak Attack: a
// delayed trigger bound to the returned object (CR 603.7a) sacrifices it at
// the beginning of the next end step, and only if you still control that
// same object (CR 400.7: a creature that left and came back is new). The
// Necromancer is sacrificed as a cost, so it is already gone when the ability
// resolves; a target that has left the graveyard makes the ability do
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1a6aa5f-2ad6-4bc7-82f0-4262f66af730",
		Name:         "Apprentice Necromancer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, {T}, Sacrifice this creature: Return target creature card from your graveyard to the battlefield. That creature gains haste. At the beginning of the next end step, sacrifice it.",
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreatureInYourGraveyard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				obj, ok, err := effects.S82OldB233ReturnWithHaste(ctx, t.ID)
				if err != nil || !ok {
					return err
				}
				return effects.ScheduleDelayedTrigger{
					Label:  "Apprentice Necromancer — sacrifice the returned creature",
					Body:   effects.StdB66MeekAttackSacrifice,
					Params: game.EffectParams{Object: obj},
				}.Apply(ctx)
			},
		}},
	})
}

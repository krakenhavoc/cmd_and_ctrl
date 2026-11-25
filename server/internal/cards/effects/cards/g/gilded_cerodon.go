package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gilded Cerodon — Creature — Beast {4}{R}, 4/4 (Hour of Devastation, S82 OLD
// batch 241, #5534):
//
//	"Whenever this creature attacks, if you control a Desert or there is a
//	 Desert card in your graveyard, target creature can't block this turn."
//
// An attack trigger with an intervening "if" (CR 603.4), checked as it attacks
// and again as it resolves; the target is any creature, chosen as the trigger
// goes on the stack.
//
// No simplification.
func init() {
	t := effects.InterveningIfYou("you control a Desert or there is a Desert card in your graveyard",
		effects.OldB241ControlsDesertOrHasOneInGraveyard,
		effects.On(game.EventAttack, effects.ThisAttacked, "Gilded Cerodon — target creature can't block this turn",
			func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, tg := range ctx.LegalTargets() {
					if tg.Kind != game.TargetCard {
						continue
					}
					if err := (effects.RestrictUntilEOT{Target: tg.ID, Restrictions: game.CantBlock, Label: "Gilded Cerodon — can't block"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}))
	t.Targets = effects.TargetCreature("target creature")
	effects.Register(effects.Spec{
		OracleID:     "93dd5f2e-aa13-4616-bfe6-466524b274ff",
		Name:         "Gilded Cerodon",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

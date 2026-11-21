package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sneak Attack — Enchantment {3}{R} (Enchanting Tales):
//
//	"{R}: You may put a creature card from your hand onto the battlefield.
//	 That creature gains haste. Sacrifice the creature at the beginning of
//	 the next end step."
//
// Meek Attack without the size limit and at {R}: Quicksilver Amulet's
// optional put from hand, then a tail that runs once the card has landed. The
// creature gains haste until end of turn, and a delayed trigger bound to that
// object sacrifices it at the beginning of the next end step (CR 603.7a),
// only if you still control that same object (a creature that left and came
// back is a new object). Declining the put makes no delayed trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7aee670-e6ac-4b3a-b3d1-7b7013db8f3d",
		Name:         "Sneak Attack",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: You may put a creature card from your hand onto the battlefield. That creature gains haste. Sacrifice the creature at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PutFromHandOntoBattlefield{
					Match:    effects.Creature(),
					Optional: true,
					Label:    "Sneak Attack — you may put a creature card from your hand onto the battlefield",
					ThenTail: sneakAttackEntered,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// sneakAttackEntered is "that creature gains haste. Sacrifice the creature at
// the beginning of the next end step". An on-disk key.
var sneakAttackEntered = effects.PutFromHandTail("sneak-attack/entered", func(ctx *effects.Context, res effects.PutFromHandResult) error {
	if res.Entered == uuid.Nil {
		return nil
	}
	c, ok := effects.BattlefieldCardForEffect(ctx.Game, res.Entered)
	if !ok {
		return nil
	}
	if err := (effects.GrantKeywordUntilEOT{Target: res.Entered, Keywords: []string{"haste"},
		Label: "Sneak Attack — haste"}).Apply(ctx); err != nil {
		return err
	}
	return effects.ScheduleDelayedTrigger{
		Label:  "Sneak Attack — sacrifice the creature",
		Body:   effects.StdB66MeekAttackSacrifice,
		Params: game.EffectParams{Object: game.ObjectRef{ID: res.Entered, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})

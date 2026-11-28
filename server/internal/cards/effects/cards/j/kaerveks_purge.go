package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kaervek's Purge — Sorcery {X}{B}{R} (Mirage):
//
//	"Destroy target creature with mana value X. If that creature dies this way,
//	 Kaervek's Purge deals damage equal to the creature's power to the
//	 creature's controller."
//
// X is announced with the spell and the target must have exactly that mana
// value, checked at announce and again at resolution (CR 608.2b). The
// creature's power and controller are read before it is destroyed (CR 608.2h);
// the damage is dealt only if it actually went to a graveyard (an indestructible
// or regenerated creature takes nothing with it), as Saw in Half reads "dies".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "daa86c76-aacb-4244-9639-81f758c634b2",
		Name:         "Kaervek's Purge",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with mana value X").WithManaValueEqualsX(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			controller, power := victim.Controller, victim.CurrentPower()
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: controller, Amount: power}.Apply(ctx)
		},
	})
}

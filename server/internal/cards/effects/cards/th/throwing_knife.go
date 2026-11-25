package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Throwing Knife — Artifact — Equipment {2} (Magic Origins, S82 older sets):
//
//	"Equipped creature gets +2/+0.
//	 Whenever equipped creature attacks, you may sacrifice this Equipment. If you
//	 do, this Equipment deals 2 damage to any target.
//	 Equip {2}"
//
// Saw's attack trigger with Servant of the Stinger's "you may sacrifice this. If
// you do": the target is chosen as the trigger goes on the stack, before the
// choice to sacrifice (CR 603.3d), the question is asked as it resolves, and the
// damage is dealt only if the Knife really was sacrificed (CR 701.21a), by the
// Knife as it last existed (CR 608.2h). A Knife that is no longer on the
// battlefield is not sacrificed and deals nothing; a target that became illegal
// is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3973121-ed59-48b3-8efa-754a1b2c704d",
		Name:         "Throwing Knife",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Optional(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return source.IsAttachedTo(ev.CardID)
				}, "Throwing Knife — you may sacrifice this Equipment. If you do, it deals 2 damage to any target", throwingKnifeSacrifice),
				"Throwing Knife — sacrifice this Equipment to deal 2 damage to the target?"),
				effects.TargetAny()),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

// throwingKnifeSacrifice is "sacrifice this Equipment. If you do, it deals 2
// damage to any target". An on-disk key: never renamed, never reused.
var throwingKnifeSacrifice = effects.SacrificeThisThen("then/s82-old-b323-throwing-knife-damage", func(ctx *effects.Context) error {
	t, ok := effects.FirstLegalTarget(ctx)
	if !ok {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
})

package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arachnus Web — Enchantment — Aura {2}{G} (S82 GAP, #5948):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block, and its activated abilities
//	 can't be activated.
//	 At the beginning of the end step, if enchanted creature's power is 4 or
//	 greater, destroy this Aura."
//
// Prison Sentence's restriction; the trigger is each end step's (any player's),
// with the intervening "if" checked as the step begins and as it resolves.
func init() {
	const label = "Arachnus Web — destroy this Aura"
	effects.Register(effects.Spec{
		OracleID:     "8b25f353-0798-4a03-a928-6a701a919468",
		Name:         "Arachnus Web",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana),
		},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "enchanted creature's power is 4 or greater",
				Test: arachnusWebHostPowerAtLeastFour,
			}, effects.On(game.EventBeginEndStep, effects.AnyPlayer, label, effects.DestroyThisPermanent(false))),
		},
	})
}

func arachnusWebHostPowerAtLeastFour(ctx *effects.Context) bool {
	host := effects.AttachedHostFor(ctx.Game, ctx.Source())
	return host != nil && effects.B31CurrentPowerOf(ctx.Game, host.InstanceID) >= 4
}

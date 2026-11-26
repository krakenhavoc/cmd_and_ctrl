package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ice Cage — Enchantment — Aura {1}{U} (Magic 2012, S82 older sets batch 451,
// #5573):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block, and its activated abilities
//	 can't be activated.
//	 When enchanted creature becomes the target of a spell or ability,
//	 destroy this Aura."
//
// Arrest's restriction. The trigger watches the targeting event for the
// creature the Aura is attached to; the targeting spell or ability still
// resolves, the Aura just being gone by then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e37416cc-f343-44cd-8676-4ce4d4068e85",
		Name:         "Ice Cage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.AttachedTo.Kind == game.TargetCard && ev.CardID == source.AttachedTo.ID
			}, "Ice Cage — destroy this Aura",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					return effects.DestroyTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

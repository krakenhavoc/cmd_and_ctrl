package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Captured by the Consulate — Enchantment — Aura {3}{W} (Kaladesh, #5539):
//
//	"Enchant creature you don't control
//	 Enchanted creature can't attack.
//	 Whenever an opponent casts a spell, if it has a single target, change the
//	 target to enchanted creature if able."
//
// "Enchant creature you don't control" is Honest Work's clause (any creature
// another player controls). The restriction is Bound in Silence's, without the
// block half. The trigger is Hydroelectric Specimen's pinned retarget (CR
// 115.7a, #1743) with the Aura's host as the destination: it watches an
// opponent's cast of a spell with exactly one target (the intervening "if",
// checked as the spell is cast and again as the trigger resolves, CR 603.4)
// and is mandatory, so the engine changes the target to the enchanted creature
// whenever that creature is a legal target for it under the spell's own
// clause, hexproof and protection judged for the spell's controller, and
// changes nothing when it is not (the "if able"). The enchanted creature is the
// one the Aura is attached to as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da164485-c985-42fa-819e-42f0d33237c5",
		Name:         "Captured by the Consulate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.OpponentControls()),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttack),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.HasASingleTarget()),
				"Captured by the Consulate — change the target of that spell to enchanted creature if able",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					spell := ctx.Trigger().Event.CardID
					if g.StackItemTargetCountForEffect(spell) != 1 {
						return nil
					}
					host, ok := effects.EnchantedCreatureAsItResolves(ctx)
					if !ok {
						return nil
					}
					return effects.ChangeTargets{
						StackID: spell,
						Policy:  game.RetargetChangeOne,
						To:      game.TargetRef{Kind: game.TargetCard, ID: host},
						Reason:  "Captured by the Consulate — change the target to the enchanted creature",
					}.Apply(ctx)
				}),
		},
	})
}

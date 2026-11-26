package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crystallization — Enchantment — Aura {G/U}{W} (Alara Reborn):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block.
//	 When enchanted creature becomes the target of a spell or ability, exile
//	 that creature."
//
// Ice Cage's trigger shape: it watches the targeting event for the creature
// the Aura is attached to, and the exile is of that creature as it still
// is when the trigger resolves (CR 400.7); one that has left is not
// touched. The spell or ability that targeted it then has no legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e50d4c9-343b-42e7-b777-9d479e97e776",
		Name:         "Crystallization",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.AttachedTo.Kind == game.TargetCard && ev.CardID == source.AttachedTo.ID
			}, "Crystallization — exile that creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id := ctx.Trigger().Event.CardID
					if id == [16]byte{} || !effects.OnBattlefield(g, id) {
						return nil
					}
					if info, ok := ctx.TriggeringPermanent(); ok && info.Left {
						return nil
					}
					return effects.ExileTarget{Target: id}.Apply(ctx)
				}),
		},
	})
}

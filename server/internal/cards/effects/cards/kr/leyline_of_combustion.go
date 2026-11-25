package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leyline of Combustion — Enchantment {2}{R}{R} (Core Set 2020):
//
//	"If this card is in your opening hand, you may begin the game with it on
//	 the battlefield.
//	 Whenever you and/or at least one permanent you control becomes the target
//	 of a spell or ability an opponent controls, this enchantment deals 2
//	 damage to that player."
//
// The opening-hand clause is Spec.OpeningHand (CR 103.6a). The trigger is
// Leovold's watch on EventBecomesTarget, which the engine emits once per
// target; "and/or" is one occurrence per spell or ability (CR 603.2c), so a
// spell that targets you and your creature, or two of your creatures, triggers
// it once. The batch key is the targeting stack item. The damage goes to the
// opponent who controls that spell or ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49a5c149-5067-4840-8385-c2989d7e5a34",
		Name:         "Leyline of Combustion",
		Completeness: effects.CompletenessFull,
		OpeningHand:  effects.BeginTheGameOnTheBattlefield(),
		Triggered: []game.TriggeredAbility{
			func() game.TriggeredAbility {
				t := effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.OldB157YouOrYourPermanentIsTargeted(ev, source, g)
				},
					"Leyline of Combustion — 2 damage to that player",
					func(g *game.Game, item *game.StackItem) error {
						if item.Trigger == nil || item.Trigger.Event.Actor == uuid.Nil {
							return nil
						}
						return effects.DealDamage{Source: item.SourceCardID, Target: item.Trigger.Event.Actor, Amount: 2}.Apply(effects.NewContext(g, item))
					})
				t.OncePerBatch = true
				t.BatchKey = func(ev game.Event, _ *game.Card, _ *game.Game) string { return ev.StackItemID.String() }
				return t
			}(),
		},
	})
}

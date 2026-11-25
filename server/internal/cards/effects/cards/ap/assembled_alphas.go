package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Assembled Alphas — Creature — Wolf {5}{R}, 5/5 (Eldritch Moon, S82 OLD,
// #5541):
//
//	"Whenever this creature blocks or becomes blocked by a creature, this
//	 creature deals 3 damage to that creature and 3 damage to that
//	 creature's controller."
//
// Ornery Goblin's per-pair block trigger (CR 509.3b, 509.3d). The Alphas
// deal both damages as one object, so one that has left still deals them.
// "That creature's controller" is read as the trigger resolves, from the
// creature as it last existed if it has left the battlefield (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "151a681d-a350-4d1d-86cd-b80d58602810",
		Name:         "Assembled Alphas",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SkewerSlingerFoe(ev, source.InstanceID) != uuid.Nil
			}, "Assembled Alphas — 3 damage to that creature and 3 damage to that creature's controller",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					ref, ok := ctx.SourceRef()
					foe := effects.SkewerSlingerFoe(item.Trigger.Event, item.SourceCardID)
					if !ok || foe == uuid.Nil {
						return nil
					}
					var controller uuid.UUID
					if c, ok := g.LookupCardForEffect(foe); ok {
						if z := g.FindCardZoneForEffect(foe); z != nil && z.Kind == game.ZoneBattlefield {
							controller = c.Controller
						}
					}
					if controller == uuid.Nil {
						if info, ok := g.LastKnownPermanentForEffect(foe); ok {
							controller = info.Controller
						}
					}
					return g.DamageInstanceForEffect(func() error {
						if err := (effects.DealDamage{SourceObject: &ref, Target: foe, Amount: 3}).Apply(ctx); err != nil {
							return err
						}
						if controller == uuid.Nil {
							return nil
						}
						return effects.DealDamage{SourceObject: &ref, Target: controller, Amount: 3}.Apply(ctx)
					})
				}),
		},
	})
}

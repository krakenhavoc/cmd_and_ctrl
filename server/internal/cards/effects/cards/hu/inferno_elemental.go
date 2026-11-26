package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inferno Elemental — Creature — Elemental {4}{R}{R}, 4/4 (Magic 2010, #5585):
//
//	"Whenever this creature blocks or becomes blocked by a creature, this
//	 creature deals 3 damage to that creature."
//
// One trigger per block pair that names the Elemental (Engulfing Slagwurm's
// pair reader); the damage comes from the Elemental as it last existed
// (CR 608.2h) and goes to the other creature if it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "287cb63c-d52d-4871-83d1-792a41ae7adc",
		Name:         "Inferno Elemental",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SkewerSlingerFoe(ev, source.InstanceID) != uuid.Nil
			}, "Inferno Elemental — 3 damage to that creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					foe := effects.SkewerSlingerFoe(item.Trigger.Event, item.SourceCardID)
					if foe == uuid.Nil || !effects.OnBattlefield(g, foe) {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: foe, Amount: 3}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

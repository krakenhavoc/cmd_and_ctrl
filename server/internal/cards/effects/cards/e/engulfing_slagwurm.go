package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Engulfing Slagwurm — {5}{G}{G} Creature — Wurm 7/7 (Jumpstart 2022, #5948):
//
//	"Whenever this creature blocks or becomes blocked by a creature, destroy
//	 that creature. You gain life equal to that creature's toughness."
//
// One trigger per block pair that names the Slagwurm (Ornery Goblin's
// pair reader). The toughness is read before the destruction as the
// creature stands, and the life is gained whether or not the creature
// is destroyed (a regeneration shield or indestructible does not undo
// "you gain life").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0340cfcb-fc96-40d2-8175-31a0093615bb",
		Name:         "Engulfing Slagwurm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SkewerSlingerFoe(ev, source.InstanceID) != uuid.Nil
			}, "Engulfing Slagwurm — destroy that creature and gain life equal to its toughness",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					foe := effects.SkewerSlingerFoe(item.Trigger.Event, item.SourceCardID)
					if foe == uuid.Nil || !effects.OnBattlefield(g, foe) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					gain := effects.DepartedCreatureToughness(g, foe)
					if err := (effects.DestroyTarget{Target: foe}).Apply(ctx); err != nil {
						return err
					}
					return effects.GainLife{Player: item.Controller, Amount: gain}.Apply(ctx)
				}),
		},
	})
}

package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Venerated Teacher — Creature — Human Wizard {2}{U}, 2/2 (Rise of the
// Eldrazi, S82 OLD batch 482, #5581):
//
//	"When this creature enters, put two level counters on each creature you
//	 control with level up."
//
// "Level up" is read through the layered view (a creature that lost its
// abilities does not have it). The set is fixed as the ability resolves and
// each creature gets its two counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be118a2d-a27e-4174-86db-d9a2d304da2f",
		Name:         "Venerated Teacher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Venerated Teacher — put two level counters on each creature you control with level up",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller != item.Controller || !c.IsCreature() || !effects.OldB482HasLevelUp(c) {
							continue
						}
						if err := (effects.AddCounter{Target: c.InstanceID, Kind: game.CounterLevel, N: 2}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}

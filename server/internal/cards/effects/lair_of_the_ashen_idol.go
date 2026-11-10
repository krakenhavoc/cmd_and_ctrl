package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lair of the Ashen Idol — Plane — Azgol (#4870):
//
//	"At the beginning of your upkeep, sacrifice a creature. If you can't,
//	 planeswalk.
//	 Whenever chaos ensues, any number of target players each create a 2/2
//	 black Zombie creature token."
//
// The sacrifice is the planar controller's own choice and mandatory (the
// ruling); with no creature they can sacrifice they planeswalk instead (CR
// 701.21a: "can't" is having nothing to sacrifice, or a rule stopping it).
// The chaos ability may target zero players or any number of them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac4fe2a2-76ff-49d9-b506-26a17b4bf892",
		Name:         "Lair of the Ashen Idol",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Lair of the Ashen Idol — sacrifice a creature; if you can't, planeswalk",
				func(g *game.Game, item *game.StackItem) error {
					var creatures []uuid.UUID
					if g.CanEffectCauseSacrificeForEffect(item.Controller) {
						creatures = b04CreatureIDsControlledBy(g, item.Controller)
					}
					if len(creatures) == 0 {
						if g.Planechase == nil {
							return nil
						}
						return g.PlaneswalkForEffect(item.Controller)
					}
					return SacrificeChoice{
						Player: item.Controller, Candidates: creatures,
						Question: "Lair of the Ashen Idol — sacrifice a creature",
					}.Apply(NewContext(g, item))
				}),
			Targeting(WheneverChaosEnsues("Lair of the Ashen Idol — any number of target players each create a 2/2 black Zombie creature token",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						if err := (CreateToken{Controller: t.ID, Template: TokenCard("2/2 black Zombie"), N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
				TargetPlayer("any number of target players").WithCount(0, 0)),
		},
	})
}

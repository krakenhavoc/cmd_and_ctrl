package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Githzerai Monk — Creature — Gith Monk {4}{W}, 3/2 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Flash
//	 Flying
//	 Psychic Defense — When this creature enters, tap all creatures you don't
//	 control."
//
// Every creature on the battlefield its controller does not control, taken as
// the trigger resolves, is tapped as one group (Benefactor's Draught's loop
// over the battlefield).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d4d3cba8-562d-4208-ab0a-c82b4dd8e959",
		Name:            "Githzerai Monk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Githzerai Monk — tap all creatures you don't control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, c := range g.BattlefieldCardsForEffect() {
						if !c.IsCreature() || c.Controller == item.Controller {
							continue
						}
						if err := (effects.TapTarget{Target: c.InstanceID}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}

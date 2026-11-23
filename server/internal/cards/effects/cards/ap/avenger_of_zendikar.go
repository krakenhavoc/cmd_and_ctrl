package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avenger of Zendikar — Creature — Elemental {5}{G}{G}, 5/5 (EDHREC
// rank 290):
//
//	"When this creature enters, create a 0/1 green Plant creature
//	 token for each land you control.
//	 Landfall — Whenever a land you control enters, you may put a
//	 +1/+1 counter on each Plant creature you control."
//
// Green's seven-mana army in a can. Two triggers, both real:
//
//   - The ETB counts the lands its controller controls at
//     resolution and makes that many Plants — so a land that entered
//     in response counts, as printed.
//   - Landfall is an ETB-filtered trigger with a "you may" prompt;
//     on yes, every creature with the Plant subtype the controller
//     controls (post-layer subtype, so any Plant, not just the
//     Avenger's own) gets a +1/+1 counter. Not targeted, so hexproof
//     and the like never matter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ba5b3f6-503b-43e6-b66e-4f8c55cffed7",
		Name:         "Avenger of Zendikar",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Avenger of Zendikar — a Plant for each land you control", func(g *game.Game, item *game.StackItem) error {
				n := effects.B02CountLandsControlledBy(g, item.Controller)
				if n == 0 {
					return nil
				}
				return effects.CreateToken{
					Controller: item.Controller,
					Template:   effects.TokenCard("0/1 green Plant"),
					N:          n,
				}.Apply(effects.NewContext(g, item))
			}),
			effects.Optional(effects.Landfall("Avenger of Zendikar — +1/+1 counter on each Plant (landfall)", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, id := range effects.B02PlantsYouControl(g, item.Controller, g.TextReaderForEffect(item).TextSubtype("Plant")) {
					if err := (effects.AddCounter{Target: id, Kind: "+1/+1", N: 1}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}), "Avenger of Zendikar — put a +1/+1 counter on each Plant you control?"),
		},
	})
}

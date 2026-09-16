package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "4ba5b3f6-503b-43e6-b66e-4f8c55cffed7",
		Name:         "Avenger of Zendikar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Avenger of Zendikar — a Plant for each land you control", func(g *game.Game, item *game.StackItem) error {
				n := b02CountLandsControlledBy(g, item.Controller)
				if n == 0 {
					return nil
				}
				return CreateToken{
					Controller: item.Controller,
					Template:   b02PlantToken(),
					N:          n,
				}.Apply(NewContext(g, item))
			}),
			Optional(Landfall("Avenger of Zendikar — +1/+1 counter on each Plant (landfall)", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, id := range b02PlantsYouControl(g, item.Controller) {
					if err := (AddCounter{Target: id, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}), "Avenger of Zendikar — put a +1/+1 counter on each Plant you control?"),
		},
	})
}

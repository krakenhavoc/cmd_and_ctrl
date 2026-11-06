package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glimmer Seeker — Creature — Human Survivor {2}{W}, 3/3 (Duskmourn:
// House of Horror, #4466):
//
//	"Survival — At the beginning of your second main phase, if this
//	 creature is tapped, draw a card if you control a Glimmer creature. If
//	 you don't control a Glimmer creature, create a 1/1 white Glimmer
//	 enchantment creature token."
//
// Defiant Survivor's Survival trigger (survival.go), its "if" asked as
// the phase begins and again as it resolves. Which half happens is read as
// it resolves: a Glimmer creature you control, read through the layers
// and in the card's own text (controlsAGlimmerCreature, Fear of the
// Dark's), means a card; none means Grand Entryway's Glimmer token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cbb5526b-3aa1-43d9-acdb-bd230ce3fe63",
		Name:         "Glimmer Seeker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Survival("Glimmer Seeker — draw a card if you control a Glimmer creature, otherwise create a Glimmer",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if controlsAGlimmerCreature(g, item.Controller, ctx.TextSubtype("Glimmer")) {
						return DrawCards{N: 1}.Apply(ctx)
					}
					return CreateToken{Template: TokenCard("1/1 white Glimmer enchantment"), N: 1}.Apply(ctx)
				}),
		},
	})
}

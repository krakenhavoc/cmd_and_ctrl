package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Earth-Origin Yak (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	When this creature enters, creatures you control get +1/+1 until end of turn.
//
// {3}{W} Creature — Ox 2/4. Inspiring Captain's enter trigger; the creatures
// are read as it resolves, so the Yak itself is pumped too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2b659ef-68f6-452c-94b5-2872f4e7b069",
		Name:         "Earth-Origin Yak",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Earth-Origin Yak — creatures you control get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl()),
						Power: 1, Toughness: 1,
						Label: "Earth-Origin Yak — +1/+1",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

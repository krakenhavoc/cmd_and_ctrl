package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian of Pilgrims — Creature — Spirit Cleric {1}{W}, 2/2 (Innistrad
// Remastered):
//
//	"When this creature enters, target creature gets +1/+1 until end of turn."
//
// The target may be this creature itself.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a090bc9c-d253-4a29-87d6-aa849f856d36",
		Name:         "Guardian of Pilgrims",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Guardian of Pilgrims — target creature gets +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1,
							Label: "Guardian of Pilgrims — +1/+1"}.Apply(ctx)
					}
					return nil
				}), effects.TargetCreature("target creature")),
		},
	})
}

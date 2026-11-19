package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guardian of Pilgrims — Creature — Spirit Cleric {1}{W}, 2/2 (Innistrad
// Remastered):
//
//	"When this creature enters, target creature gets +1/+1 until end of turn."
//
// The target may be this creature itself.
func init() {
	Register(Spec{
		OracleID:     "a090bc9c-d253-4a29-87d6-aa849f856d36",
		Name:         "Guardian of Pilgrims",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Guardian of Pilgrims — target creature gets +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1,
							Label: "Guardian of Pilgrims — +1/+1"}.Apply(ctx)
					}
					return nil
				}), TargetCreature("target creature")),
		},
	})
}

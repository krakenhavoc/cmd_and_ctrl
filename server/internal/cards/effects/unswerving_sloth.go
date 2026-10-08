package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unswerving Sloth — Creature — Sloth Mount {3}{W}{W}:
//
//	"Whenever this creature attacks while saddled, it gains indestructible
//	 until end of turn. Untap all creatures you control.
//	 Saddle 4"
//
// The untap includes the creatures that saddled it, which is the point
// of the card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5e512ce9-5e2e-4ec3-9486-a10c63cff299",
		Name:         "Unswerving Sloth",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Saddle(4)},
		Triggered: []game.TriggeredAbility{
			AttacksWhileSaddled("Unswerving Sloth — indestructible until end of turn; untap all creatures you control", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"indestructible"}, Label: "Unswerving Sloth — indestructible until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return UntapAllCreaturesYouControl{}.Apply(ctx)
			}),
		},
	})
}

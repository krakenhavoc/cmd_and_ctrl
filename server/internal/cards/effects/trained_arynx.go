package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trained Arynx — Creature — Cat Beast Mount {1}{W}:
//
//	"Whenever this creature attacks while saddled, it gains first strike
//	 until end of turn. Scry 1.
//	 Saddle 2"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc875500-f6df-4365-b2d6-7cf425daea97",
		Name:         "Trained Arynx",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Saddle(2)},
		Triggered: []game.TriggeredAbility{
			AttacksWhileSaddled("Trained Arynx — first strike until end of turn; scry 1", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"first strike"}, Label: "Trained Arynx — first strike until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return Scry{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		},
	})
}

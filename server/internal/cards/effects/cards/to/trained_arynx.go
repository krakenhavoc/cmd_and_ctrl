package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trained Arynx — Creature — Cat Beast Mount {1}{W}:
//
//	"Whenever this creature attacks while saddled, it gains first strike
//	 until end of turn. Scry 1.
//	 Saddle 2"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc875500-f6df-4365-b2d6-7cf425daea97",
		Name:         "Trained Arynx",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Saddle(2)},
		Triggered: []game.TriggeredAbility{
			effects.AttacksWhileSaddled("Trained Arynx — first strike until end of turn; scry 1", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"first strike"}, Label: "Trained Arynx — first strike until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return effects.Scry{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		},
	})
}

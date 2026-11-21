package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Person of Interest — Creature — Human Rogue {3}{R}, 3/3:
//
//	"When this creature enters, suspect it. Create a 2/2 white and blue
//	 Detective creature token. (A suspected creature has menace and
//	 can't block.)"
//
// Suspect (CR 701.60, #2698) then the token, in printed order. The
// Detective is a plain 2/2: only Person of Interest is suspected.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d18cd9b-6013-48d7-9269-4787e7585a51",
		Name:         "Person of Interest",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Person of Interest — suspect it, then create a Detective", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.Suspect{Target: item.SourceCardID}).Apply(ctx); err != nil {
					return err
				}
				return effects.CreateToken{Template: effects.TokenCard("2/2 white and blue Detective"), N: 1}.Apply(ctx)
			}),
		},
	})
}

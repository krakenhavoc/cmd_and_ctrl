package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Person of Interest — Creature — Human Rogue {3}{R}, 3/3:
//
//	"When this creature enters, suspect it. Create a 2/2 white and blue
//	 Detective creature token. (A suspected creature has menace and
//	 can't block.)"
//
// Suspect (CR 701.60, #2698) then the token, in printed order. The
// Detective is a plain 2/2: only Person of Interest is suspected.
func init() {
	Register(Spec{
		OracleID:     "3d18cd9b-6013-48d7-9269-4787e7585a51",
		Name:         "Person of Interest",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Person of Interest — suspect it, then create a Detective", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (Suspect{Target: item.SourceCardID}).Apply(ctx); err != nil {
					return err
				}
				return CreateToken{Template: TokenCard("2/2 white and blue Detective"), N: 1}.Apply(ctx)
			}),
		},
	})
}

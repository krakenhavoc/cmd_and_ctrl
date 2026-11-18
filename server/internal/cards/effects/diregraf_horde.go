package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diregraf Horde — Creature — Zombie {4}{B}, 3/4 (Midnight Hunt):
//
//	"When this creature enters, create two 2/2 black Zombie creature
//	 tokens with decayed. When you do, exile up to two target cards from
//	 graveyards."
//
// The second sentence is a reflexive trigger (CR 603.12): it happens
// once the tokens were made, goes on the stack above the first ability,
// and chooses its targets then. With no card to pick it is removed
// (CR 603.3d), which leaves the tokens as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d7947858-f11d-438c-8ad3-991afe950a11",
		Name:         "Diregraf Horde",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Diregraf Horde — create two 2/2 black Zombie creature tokens with decayed, then exile up to two target cards from graveyards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (CreateToken{Template: TokenCard("2/2 black Zombie with decayed"), N: 2}).Apply(ctx); err != nil {
						return err
					}
					return WhenYouDo("Diregraf Horde — exile up to two target cards from graveyards", diregrafHordeExile).Apply(ctx)
				}),
		},
	})
}

// diregrafHordeExile is the reflexive half. An on-disk key: never
// renamed, never reused.
var diregrafHordeExile = game.ReflexiveBody("diregraf-horde/exile", simpleBody(ExileTargetCards),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("up to two target cards from graveyards").WithCount(0, 2)
	}))

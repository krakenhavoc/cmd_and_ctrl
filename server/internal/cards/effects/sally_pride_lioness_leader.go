package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sally Pride, Lioness Leader — {3}{W}{W} Legendary Creature — Cat Mutant
// Rebel 2/4 (Teenage Mutant Ninja Turtles, #4457):
//
//	"When Sally Pride enters, create X 2/2 red Mutant creature tokens,
//	 where X is the number of nontoken creatures you control.
//	 Whenever Sally Pride attacks, put a +1/+1 counter on each creature
//	 you control."
//
// X is counted as the enters trigger resolves (CR 608.2h), Sally included if
// she is still there; a Mutant made by an earlier trigger is a token and is
// not counted. The attack trigger puts one counter on each creature its
// controller controls as it resolves, Sally and the attackers included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b529f6b0-c6b2-4eb3-bb48-178e1eda75f3",
		Name:         "Sally Pride, Lioness Leader",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sally Pride — create a 2/2 red Mutant for each nontoken creature you control",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && !c.IsToken() {
							n++
						}
					}
					if n <= 0 {
						return nil
					}
					return CreateToken{Controller: item.Controller, Template: TokenCard("2/2 red Mutant"), N: n}.Apply(NewContext(g, item))
				}),
			WheneverThisAttacks("Sally Pride — put a +1/+1 counter on each creature you control", putACounterOnEachCreatureYouControl(game.CounterPlusOne)),
		},
	})
}

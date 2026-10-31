package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bold Biochemist — {1}{U} Creature — Human Scientist 1/3 (Marvel Super
// Heroes, #4455):
//
//	"Power-up — {5}{U}: Put a +1/+1 counter on this creature and draw two
//	 cards. (Activate each power-up ability only once. Reduce the cost by
//	 its mana cost if it entered this turn.)"
//
// Power-up is the engine's (CR 702.193, game/power_up.go): the turn it
// enters the ability costs {5}{U} less {1}{U}, which is {4}; after that
// {5}{U}. The counter and the draw are one resolution; the counter is
// skipped if the creature has left, the draw is not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "45ddfa35-7b3c-4916-b231-dc721b7b56d9",
		Name:         "Bold Biochemist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			powerUpAnswering("{5}{U}: Put a +1/+1 counter on this creature and draw two cards.",
				ManaCost("{5}{U}"), func(g *game.Game, item *game.StackItem) error {
					if err := plusOneCountersOnThis(1)(g, item); err != nil {
						return err
					}
					return DrawCards{N: 2}.Apply(NewContext(g, item))
				}),
		},
	})
}

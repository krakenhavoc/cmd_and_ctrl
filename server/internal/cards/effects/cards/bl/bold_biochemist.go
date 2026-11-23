package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "45ddfa35-7b3c-4916-b231-dc721b7b56d9",
		Name:         "Bold Biochemist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.PowerUpAnswering("{5}{U}: Put a +1/+1 counter on this creature and draw two cards.",
				effects.ManaCost("{5}{U}"), func(g *game.Game, item *game.StackItem) error {
					if err := effects.PlusOneCountersOnThis(1)(g, item); err != nil {
						return err
					}
					return effects.DrawCards{N: 2}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

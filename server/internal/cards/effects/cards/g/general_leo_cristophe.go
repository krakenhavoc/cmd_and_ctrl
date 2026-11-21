package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// General Leo Cristophe — Legendary Creature — Human Soldier {4}{W}, 2/2
// (deck request #3655):
//
//	"When General Leo Cristophe enters, return up to one target creature
//	 card with mana value 3 or less from your graveyard to the
//	 battlefield. Then put a +1/+1 counter on General Leo Cristophe for
//	 each creature you control."
//
// The count is taken after the return, so the returned creature and Leo
// himself both count.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("General Leo Cristophe — return up to one creature card with mana value 3 or less, then +1/+1 counters for each creature you control",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			if err := effects.ReturnFirstLegalGraveyardTargetToBattlefield(g, item); err != nil {
				return err
			}
			n := effects.CreaturesControlledBy(g, item.Controller)
			if n == 0 {
				return nil
			}
			return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
		})
	enters.Targets = effects.TargetCardInGraveyard("up to one target creature card with mana value 3 or less from your graveyard",
		effects.YouOwn(), effects.Creature(), effects.ManaValueLE(3)).WithCount(0, 1)
	effects.Register(effects.Spec{
		OracleID:     "088166b8-5b0b-4ba9-9b0d-358ae523f336",
		Name:         "General Leo Cristophe",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}

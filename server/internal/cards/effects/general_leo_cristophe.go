package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	enters := WhenThisEnters("General Leo Cristophe — return up to one creature card with mana value 3 or less, then +1/+1 counters for each creature you control",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			if err := returnFirstLegalGraveyardTargetToBattlefield(g, item); err != nil {
				return err
			}
			n := creaturesControlledBy(g, item.Controller)
			if n == 0 {
				return nil
			}
			return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
		})
	enters.Targets = TargetCardInGraveyard("up to one target creature card with mana value 3 or less from your graveyard",
		YouOwn(), Creature(), ManaValueLE(3)).WithCount(0, 1)
	Register(Spec{
		OracleID:     "088166b8-5b0b-4ba9-9b0d-358ae523f336",
		Name:         "General Leo Cristophe",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}

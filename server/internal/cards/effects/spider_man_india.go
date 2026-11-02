package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Man India — Legendary Creature — Spider Human Hero {3}{G}{W}, 4/4
// (Marvel's Spider-Man, #4460):
//
//	"Web-slinging {1}{G}{W} (You may cast this spell for {1}{G}{W} if you also
//	 return a tapped creature you control to its owner's hand.)
//	 Pavitr's Seva — Whenever you cast a creature spell, put a +1/+1 counter on
//	 target creature you control. It gains flying until end of turn."
//
// Web-slinging is Amazing Spider-Girl's alternative cost. The cast trigger
// targets as it goes on the stack and does both halves to that one creature
// as it resolves (nothing happens if the target is gone, CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8933297c-62f1-4df9-91b8-2d3481db77e5",
		Name:         "Spider-Man India",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:          "web_slinging",
			Label:        "Web-slinging {1}{G}{W}",
			ManaCost:     "{1}{G}{W}",
			ReturnToHand: PermanentYouControl("a tapped creature you control", Creature(), b751Tapped()),
			PayLabel:     "a tapped creature you control",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverYouCast(Creature(), "Pavitr's Seva — put a +1/+1 counter on target creature you control; it gains flying until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok || !onBattlefield(g, id) {
						return nil
					}
					if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Pavitr's Seva — flying until end of turn"}.Apply(ctx)
				}),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

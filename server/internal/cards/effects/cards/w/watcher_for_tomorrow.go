package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Watcher for Tomorrow — Creature — Human Wizard {1}{U}, 2/1 (Modern
// Horizons):
//
//	"Hideaway 4 (When this creature enters, look at the top four cards of your
//	 library, exile one face down, then put the rest on the bottom in a random
//	 order.)
//	 This creature enters tapped.
//	 When this creature leaves the battlefield, put the exiled card into its
//	 owner's hand."
//
// The hideaway is the shared Hideaway, which links the exiled card to this
// permanent OBJECT (CR 607.2a). The leaves-the-battlefield trigger captures
// that object as it goes on the stack and puts every card it still hides into
// that card's owner's hand: the exiled card, however the Watcher left, and
// nothing if the card was never hidden (an empty library) or has left exile
// since. The enters-tapped clause is a replacement (CR 614.1d).
//
// No simplification.
const watcherForTomorrowLabel = "Watcher for Tomorrow — put the exiled card into its owner's hand"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "277aafa6-4270-4d24-a888-cf9e8e635d4b",
		Name:         "Watcher for Tomorrow",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			effects.Hideaway("Watcher for Tomorrow", 4),
			effects.On(game.EventLTB, effects.Self, watcherForTomorrowLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ref, ok := ctx.SourceRef()
				if !ok {
					return nil
				}
				for _, id := range g.HiddenCardsForEffect(game.PermissionCardRef(ref)) {
					if err := (effects.BounceToHand{Target: id}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}

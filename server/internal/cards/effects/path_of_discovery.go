package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Path of Discovery — Enchantment {3}{G}:
//
//	"Whenever a creature you control enters, it explores. (Reveal the
//	 top card of your library. Put that card into your hand if it's a
//	 land. Otherwise, put a +1/+1 counter on the creature, then put the
//	 card back or put it into your graveyard.)"
//
// The creature that entered explores (CR 701.44) under the Path's
// controller, through the same Explores action the Map token uses. A
// creature that has left by the time the trigger resolves still
// explores: the card is revealed and a nonland card may still go to the
// graveyard, but no counter is put on anything (CR 701.44c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "52714c3c-6821-44dd-88ba-813554df9914",
		Name:         "Path of Discovery",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, CreatureEnteredUnderYourControl, "Path of Discovery — it explores",
				func(g *game.Game, item *game.StackItem) error {
					if o := item.Trigger.Object; o != nil {
						// The object that entered, not a later one (CR 400.7).
						return g.ExploreForEffect(item.SourceCardID, game.ObjectRef{ID: o.ID, Epoch: o.Epoch}, item.Controller, nil)
					}
					return Explores{Explorer: item.Trigger.Event.CardID}.Apply(NewContext(g, item))
				}),
		},
	})
}

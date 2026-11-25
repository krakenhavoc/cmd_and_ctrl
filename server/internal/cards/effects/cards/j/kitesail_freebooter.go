package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kitesail Freebooter — Creature — Human Pirate {1}{B}, 1/2 (Core Set 2021,
// #5511):
//
//	"Flying
//	 When this creature enters, target opponent reveals their hand. You choose
//	 a noncreature, nonland card from it. Exile that card until this creature
//	 leaves the battlefield."
//
// The reveal and pick are Thought-Knot Seer's (the whole table sees the hand;
// a hand with no such card is revealed and nothing is exiled). The exile is
// linked to the Freebooter as the object it is (CR 607.2a), and the card is
// returned to its owner's hand when that object leaves the battlefield
// (CR 610.3): a Freebooter that left and came back is a new object with no
// claim on it (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8aa1bbc9-b047-4d68-93f4-728e40dafc5e",
		Name:            "Kitesail Freebooter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Kitesail Freebooter — target opponent reveals their hand; you choose a noncreature, nonland card from it and exile it until this creature leaves the battlefield",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						return effects.ChooseFromRevealedHand{
							Player: effects.TargetedPlayer(ctx),
							Filter: effects.And(effects.Nonland(), effects.Not(effects.Creature())),
							Label:  "noncreature, nonland card",
							Exile:  true,
							Linked: true,
							Then:   kitesailFreebooterReturn,
						}.Apply(ctx)
					}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

// kitesailFreebooterReturn records that the exiled card goes back to its
// owner's hand when the Freebooter leaves the battlefield. The exile linked
// the card to the Freebooter object that exiled it, which is the object the
// return waits on. An on-disk key: never renamed, never reused.
var kitesailFreebooterReturn = effects.RevealedPickThen("revealed-pick/s82-old-b108-kitesail-freebooter-return",
	func(ctx *effects.Context, pick game.RevealedPick) error {
		if len(pick.Chosen) == 0 {
			return nil
		}
		id := pick.Chosen[0].InstanceID
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.ExiledWith.ID == uuid.Nil {
			return nil
		}
		ctx.Game.ScheduleUntilReturnForEffect(id, game.UntilReturn{
			Controller:   pick.Chooser,
			SourceCardID: pick.Source,
			Label:        "Kitesail Freebooter — the exiled card returns to its owner's hand",
			Leaves:       game.ObjectRef(c.ExiledWith),
			ToHand:       true,
		})
		return nil
	})

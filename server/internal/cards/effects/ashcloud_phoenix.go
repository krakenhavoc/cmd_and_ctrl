package effects

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashcloud Phoenix — Creature — Phoenix {2}{R}{R}, 4/1 (Murders at
// Karlov Manor Commander):
//
//	"Flying
//	 When this creature dies, return it to the battlefield face down
//	 under your control.
//	 Morph {4}{R}{R}
//	 When this creature is turned face up, it deals 2 damage to each
//	 player."
//
// The return is Yedora's and Missy's face-down reanimation
// (ReturnFromGraveyardFaceDownForEffect): the face-down state rides the
// entry, so no ETB trigger runs (CR 708.2a). It is "under YOUR control":
// the trigger's controller, not the card's owner. The permanent is a
// plain 2/2 with the default body (the card lists none), and because the
// Phoenix itself prints morph it can be turned up for {4}{R}{R}
// (CR 702.37e), which deals the damage. A Phoenix that left the graveyard
// before the trigger resolves is a new object (CR 400.7) and nothing
// happens. The damage is the Phoenix's own, to every player including
// its controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "fb2d7c3d-b68f-410e-9a73-cfcbfe702ade",
		Name:             "Ashcloud Phoenix",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Morph("{4}{R}{R}")},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Ashcloud Phoenix — return it to the battlefield face down under your control",
				func(g *game.Game, item *game.StackItem) error {
					_, err := g.ReturnFromGraveyardFaceDownForEffect(item.Trigger.Event.CardID, item.Controller, false, nil)
					if errors.Is(err, game.ErrCardNotFound) {
						return nil
					}
					return err
				}),
			WhenThisIsTurnedFaceUp("Ashcloud Phoenix — it deals 2 damage to each player",
				func(g *game.Game, item *game.StackItem) error {
					return b38DamageEachPlayer(NewContext(g, item), 2)
				}),
		},
	})
}

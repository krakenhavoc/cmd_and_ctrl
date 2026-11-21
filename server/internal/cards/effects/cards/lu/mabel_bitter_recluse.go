package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mabel, Bitter Recluse — Legendary Creature — Mouse Warlock {B}, 1/1
// (Reality Fracture):
//
//	"Deathtouch
//	 When Mabel enters, remove up to three counters from another target
//	 creature or planeswalker."
//
// The counters come off one at a time, each pick naming the kind, and
// the controller may stop early ("up to"). The target is required: with
// no other creature or planeswalker on the battlefield the trigger is
// removed from the stack (CR 603.3d).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "36924990-8e3a-434e-bcd1-f03603cb350d",
		Name:            "Mabel, Bitter Recluse",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Mabel, Bitter Recluse — remove up to three counters from another target creature or planeswalker",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.RfRemoveUpToCounters(g, item.Controller, item.SourceCardID, id, 3,
							"Mabel, Bitter Recluse — remove a counter (up to three)")
					}),
				effects.Another(effects.TargetPermanent("another target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))),
			),
		},
	})
}

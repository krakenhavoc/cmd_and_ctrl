package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "36924990-8e3a-434e-bcd1-f03603cb350d",
		Name:            "Mabel, Bitter Recluse",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Mabel, Bitter Recluse — remove up to three counters from another target creature or planeswalker",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return rfRemoveUpToCounters(g, item.Controller, item.SourceCardID, id, 3,
							"Mabel, Bitter Recluse — remove a counter (up to three)")
					}),
				Another(TargetPermanent("another target creature or planeswalker", Or(Creature(), Planeswalker()))),
			),
		},
	})
}

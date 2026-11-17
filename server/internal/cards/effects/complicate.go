package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Complicate — Instant {2}{U} (Onslaught):
//
//	"Counter target spell unless its controller pays {3}.
//	 Cycling {2}{U}
//	 When you cycle this card, you may counter target spell unless its
//	 controller pays {1}."
//
// Both halves are Spiketail Drake's counterTheTargetSpellUnlessPaid: the
// spell's controller is asked as the counter resolves, the stack waits
// for the answer, and a payment they can't fund counters it (CR 118.12a).
// Cycling is the keyword's own hand ability, and the cycle trigger fires
// from the graveyard (CR 702.29c). Its target is chosen as it goes on the
// stack, so with no spell to target it is removed (CR 603.3d), and the
// "you may" is asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc09cbc0-2d28-44c6-b9a9-b95e9e58bb1b",
		Name:         "Complicate",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return counterTheTargetSpellUnlessPaid("Complicate", "{3}")(ctx.Game, item)
		},
		Activated: []ActivatedAbility{Cycling("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					InGraveyard(On(game.EventCycle, Self, "Complicate — cycled: you may counter target spell unless its controller pays {1}",
						counterTheTargetSpellUnlessPaid("Complicate", "{1}"))),
					"Complicate — counter target spell unless its controller pays {1}?"),
				TargetSpell("target spell"),
			),
		},
	})
}

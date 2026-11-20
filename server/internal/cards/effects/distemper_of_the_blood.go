package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Distemper of the Blood — Sorcery {1}{R} (Eldritch Moon, #5541):
//
//	"Target creature gets +2/+2 and gains trample until end of turn.
//	 Madness {R} (If you discard this card, discard it into exile. When you
//	 do, cast it for its madness cost or put it into your graveyard.)"
//
// Massive Might's pump at sorcery speed, plus the one Madness string
// (CR 702.35a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d3e7cde-a4ed-428f-bd37-c6dda2104221",
		Name:         "Distemper of the Blood",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2, "trample")),
		Madness:      "{R}",
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(2, 2, "Distemper of the Blood — +2/+2 and trample", "trample")(ctx.Game, item)
		},
	})
}

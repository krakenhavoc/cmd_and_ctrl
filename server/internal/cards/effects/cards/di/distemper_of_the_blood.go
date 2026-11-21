package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4d3e7cde-a4ed-428f-bd37-c6dda2104221",
		Name:         "Distemper of the Blood",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "trample")),
		Madness:      "{R}",
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(2, 2, "Distemper of the Blood — +2/+2 and trample", "trample")(ctx.Game, item)
		},
	})
}

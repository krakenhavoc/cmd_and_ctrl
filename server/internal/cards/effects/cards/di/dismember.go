package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dismember — Instant {1}{B/P}{B/P}:
//
//	"({B/P} can be paid with either {B} or 2 life.)
//	 Target creature gets -5/-5 until end of turn."
//
// Both Phyrexian symbols are paid by the engine at cast (CR 107.4f);
// the card file has only the -5/-5. Lethal toughness is the state
// check's business, as for every other shrink effect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd74f8eb-0253-42dd-8277-186d4934da38",
		Name:         "Dismember",
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -5, -5)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			return effects.BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -5,
				Toughness: -5,
				Label:     "Dismember — -5/-5",
			}.Apply(ctx)
		},
	})
}

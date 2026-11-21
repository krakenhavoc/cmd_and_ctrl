package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Engulfing Flames — Instant {R}:
//
//	"Engulfing Flames deals 1 damage to target creature. It can't be
//	 regenerated this turn.
//	 Flashback {3}{R}"
//
// "It can't be regenerated this turn" is the spell's, not the damage's,
// so it holds even when the damage is prevented (ADR 0108 §2). Cast with
// flashback the card is exiled on leaving the stack, as printed.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "88f935ee-7dbb-4c27-8485-f56b4193fa11",
		Name:             "Engulfing Flames",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{3}{R}")},
		Targets:          effects.TargetCreature("target creature"),
		Purpose:          effects.ForTargets(effects.DamageToTarget(0, 1)),
		OnResolve:        effects.DamageFirstTargetThenNoRegen(1, nil),
	})
}

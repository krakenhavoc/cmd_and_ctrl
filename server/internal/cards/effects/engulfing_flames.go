package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:         "88f935ee-7dbb-4c27-8485-f56b4193fa11",
		Name:             "Engulfing Flames",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{R}")},
		Targets:          TargetCreature("target creature"),
		Purpose:          ForTargets(DamageToTarget(0, 1)),
		OnResolve:        damageFirstTargetThenNoRegen(1, nil),
	})
}

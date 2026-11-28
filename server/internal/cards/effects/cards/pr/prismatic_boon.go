package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prismatic Boon — Instant {X}{W}{U} (Mirage, S82 OLD b799, #5633):
//
//	"Choose a color. X target creatures gain protection from the chosen color
//	 until end of turn."
//
// X is announced with the cast and is the exact number of targets
// (CountFromX, Aether Tide's clause). The color is chosen as the spell
// resolves, by its caster (the color the creatures are protected FROM), and
// the creatures still legal then gain the protection until end of turn
// (Bathe in Light's tail). A spell whose targets are all gone does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aee5dda1-2181-40d5-9846-85d154e1d5da",
		Name:         "Prismatic Boon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetsCountedByX(effects.TargetCreature("X target creatures")),
		Purpose:      game.Purpose{Answers: game.AnswerProtect, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection"))},
		OnResolve:    effects.OldB799PrismaticBoon,
	})
}

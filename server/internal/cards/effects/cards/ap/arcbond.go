package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcbond — Instant {2}{R} (Fate Reforged):
//
//	"Choose target creature. Whenever that creature is dealt damage this turn, it
//	 deals that much damage to each other creature and each player."
//
// A repeating delayed trigger about that one object (CR 603.7b, 603.7c): each
// time it is dealt damage this turn the creature itself, as the damage source,
// deals that much damage at once to every other creature and every player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01790829-cc03-430f-a899-2f274b7dbf62",
		Name:         "Arcbond",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			return effects.ArcbondChooseTarget(ctx, id)
		})),
	})
}

package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rootgrapple — Kindred Instant — Treefolk {4}{G} (Lorwyn, S82 OLD b546,
// #5593):
//
//	"Destroy target noncreature permanent. If you control a Treefolk, draw
//	 a card."
//
// The destruction, then the draw decided as the spell resolves (CR 608.2c);
// a target that is illegal by resolution stops the whole spell (CR 608.2b).
// A changeling you control is a Treefolk (CR 702.73a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6cd8c51-dd5b-4037-b6ae-1776e3743338",
		Name:         "Rootgrapple",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target noncreature permanent", effects.Noncreature()),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.DestroyTarget{Target: id}.Apply(ctx)
			}),
			effects.S82OldB546DrawIfYouControl("Treefolk"),
		),
	})
}

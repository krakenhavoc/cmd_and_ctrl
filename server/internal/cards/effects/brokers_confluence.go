package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brokers Confluence — Instant {2}{G}{W}{U} (Streets of New Capenna
// Commander):
//
//	"Choose three. You may choose the same mode more than once.
//	 • Proliferate.
//	 • Target creature phases out.
//	 • Counter target activated or triggered ability."
//
// Mystic Confluence's shape (CR 700.2d): each occurrence runs in the order
// chosen, so proliferate twice adds two counters of each kind and two
// counter-an-ability bullets have a target each. Proliferate asks for its
// choice as it runs; the phase-out is Vodalian Illusionist's (the creature
// and everything attached, CR 702.26); the third is Stifle's clause.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9c6c9cc-2dbf-48ac-ad3a-c500e7074c62",
		Name:         "Brokers Confluence",
		Completeness: CompletenessFull,
		Modes: ChooseNRepeating("Choose three (you may choose the same mode more than once)", 3, 3,
			ModeDoing("Proliferate.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return Proliferate{}.Apply(ctx)
				}),
			ModeDoing("Target creature phases out.",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return PhaseOut{Targets: []uuid.UUID{t.ID}}.Apply(ctx)
				}),
			ModeDoing("Counter target activated or triggered ability.",
				AbilityOnStack("target activated or triggered ability"),
				CounterTheModesTarget),
		),
	})
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Uneasy Partings — Instant {3}{U} (The Hobbit):
//
//	"This spell costs {1} less to cast if it targets an attacking
//	 nontoken creature.
//	 Target creature's owner puts it on their choice of the top or bottom
//	 of their library."
//
// #4454, S76. The reduction reads the announced target
// (CostsLessIfItTargets). The tuck is Endless Detour's: the owner, not
// the caster, picks top or bottom (PutInLibraryInAnyOrder with
// LibraryPlaceTopOrBottom).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "15211eaa-d906-4b95-8942-2f49d3cb4b81",
		Name:         "Uneasy Partings",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(1, "This spell costs {1} less to cast if it targets an attacking nontoken creature.",
				hobB3AttackingNontokenCreature()),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok || c.Owner == uuid.Nil {
					continue
				}
				return PutInLibraryInAnyOrder{
					Chooser:   c.Owner,
					Cards:     []uuid.UUID{t.ID},
					Placement: game.LibraryPlaceTopOrBottom,
					Label:     "Uneasy Partings — put it on the top or bottom of your library",
				}.Apply(ctx)
			}
			return nil
		},
	})
}

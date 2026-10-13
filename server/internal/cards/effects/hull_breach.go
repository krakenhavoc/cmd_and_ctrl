package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hull Breach — Sorcery {R}{G} (EDHREC rank 2402):
//
//	"Choose one —
//	 • Destroy target artifact.
//	 • Destroy target enchantment.
//	 • Destroy target artifact and target enchantment."
//
// The Gruul two-for-one. Three modes, each with its own target
// clause. The third names two "target" words with DIFFERENT
// predicates, so it is two clauses in one statement (Clauses): slot 0
// is the artifact, slot 1 the enchantment, each checked against its
// own predicate at announce and again at resolution. One artifact
// enchantment may fill both slots (CR 115.3 forbids repeating an
// object only within one instance of "target"); it is then destroyed
// once.
func init() {
	Register(Spec{
		OracleID:     "2da232d8-580f-4116-b977-2c59cd21b5a4",
		Name:         "Hull Breach",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Destroy target artifact.", TargetPermanent("target artifact", Artifact()), DestroyTheModesTarget),
			ModeDoing("Destroy target enchantment.", TargetPermanent("target enchantment", Enchantment()), DestroyTheModesTarget),
			ModeDoing("Destroy target artifact and target enchantment.",
				Clauses(
					TargetPermanent("target artifact", Artifact()),
					TargetPermanent("target enchantment", Enchantment()),
				),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					var ids []uuid.UUID
					for slot := 0; slot < 2; slot++ {
						t, ok := ModeClauseTarget(ctx, occ, slot)
						if !ok || !onBattlefield(ctx.Game, t.ID) || slices.Contains(ids, t.ID) {
							continue
						}
						ids = append(ids, t.ID)
					}
					// Both at once, as one sentence: one destroy leg.
					return ctx.Game.DestroyPermanentsThenForEffect(ids, func(*game.Game, []uuid.UUID) error { return nil })
				}),
		),
	})
}

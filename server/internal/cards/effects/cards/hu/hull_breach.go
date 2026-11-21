package cardshu

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "2da232d8-580f-4116-b977-2c59cd21b5a4",
		Name:         "Hull Breach",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target artifact.", effects.TargetPermanent("target artifact", effects.Artifact()), effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target enchantment.", effects.TargetPermanent("target enchantment", effects.Enchantment()), effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target artifact and target enchantment.",
				effects.Clauses(
					effects.TargetPermanent("target artifact", effects.Artifact()),
					effects.TargetPermanent("target enchantment", effects.Enchantment()),
				),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					var ids []uuid.UUID
					for slot := 0; slot < 2; slot++ {
						t, ok := effects.ModeClauseTarget(ctx, occ, slot)
						if !ok || !effects.OnBattlefield(ctx.Game, t.ID) || slices.Contains(ids, t.ID) {
							continue
						}
						ids = append(ids, t.ID)
					}
					// Both at once, as one sentence: one destroy leg.
					return ctx.Game.DestroyPermanentsKeyedForEffect(ids, game.KeyedTail{})
				}),
		),
	})
}

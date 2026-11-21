package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Season of Weaving — Sorcery {4}{U}{U} (Bloomburrow, #4467):
//
//	"Choose up to five {P} worth of modes. You may choose the same mode
//	 more than once.
//	 {P} — Draw a card.
//	 {P}{P} — Choose an artifact or creature you control. Create a token
//	 that's a copy of it.
//	 {P}{P}{P} — Return each nonland, nontoken permanent to its owner's
//	 hand."
//
// A pawprint spell (CR 700.2i, #4786). The {P}{P} mode does not target: the
// artifact or creature is chosen as the mode resolves (a hexproof one may be
// picked), by a registered permanent pick, and the token copies it as it
// is then. The {P}{P}{P} mode is one simultaneous bounce of every nonland,
// nontoken permanent, whoever controls it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b473a42-a0c5-43d5-93f9-c599e963a06b",
		Name:         "Season of Weaving",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseUpToPawprints(5, true,
			effects.Pawprint(1, effects.ModeDoing("Draw a card.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DrawCards{N: 1}.Apply(ctx)
				})),
			effects.Pawprint(2, effects.ModeDoing("Choose an artifact or creature you control. Create a token that's a copy of it.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.ChoosePermanents{
						Question: "Season of Weaving — choose an artifact or creature you control; create a token copy of it",
						Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
							var out []uuid.UUID
							for _, c := range g.BattlefieldCardsForEffect() {
								if c.Controller == of && (c.IsArtifact() || c.IsCreature()) {
									out = append(out, c.InstanceID)
								}
							}
							return out, 1, 1
						},
						Keyed: seasonOfWeavingCopy,
					}.Apply(ctx)
				})),
			effects.Pawprint(3, effects.ModeWithPurpose(effects.ModeDoing("Return each nonland, nontoken permanent to its owner's hand.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					var ids []uuid.UUID
					for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Nonland(), effects.Not(effects.IsTokenPredicate()))) {
						ids = append(ids, c.InstanceID)
					}
					return effects.ReturnAllToHand{Cards: ids}.Apply(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepBounce, Partial: true}})),
		),
	})
}

// seasonOfWeavingCopy is the {P}{P} mode's token copy of the chosen
// permanent. An on-disk key: never renamed, never reused.
var seasonOfWeavingCopy = effects.RegisterPermanentsPick("prompt-run/season-of-weaving-copy", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if err := (effects.CreateTokenCopy{Copy: id, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})

package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "4b473a42-a0c5-43d5-93f9-c599e963a06b",
		Name:         "Season of Weaving",
		Completeness: CompletenessFull,
		Modes: ChooseUpToPawprints(5, true,
			Pawprint(1, ModeDoing("Draw a card.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return DrawCards{N: 1}.Apply(ctx)
				})),
			Pawprint(2, ModeDoing("Choose an artifact or creature you control. Create a token that's a copy of it.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return ChoosePermanents{
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
			Pawprint(3, ModeWithPurpose(ModeDoing("Return each nonland, nontoken permanent to its owner's hand.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					var ids []uuid.UUID
					for _, c := range MatchingBattlefield(ctx, And(Nonland(), Not(IsTokenPredicate()))) {
						ids = append(ids, c.InstanceID)
					}
					return ReturnAllToHand{Cards: ids}.Apply(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepBounce, Partial: true}})),
		),
	})
}

// seasonOfWeavingCopy is the {P}{P} mode's token copy of the chosen
// permanent. An on-disk key: never renamed, never reused.
var seasonOfWeavingCopy = RegisterPermanentsPick("prompt-run/season-of-weaving-copy", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if err := (CreateTokenCopy{Copy: id, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})

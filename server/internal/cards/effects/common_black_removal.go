package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Common Black Removal — Instant {2}{B}{B}:
//
//	"Choose one. Destroy target creature, then —
//	 • Create a Food token.
//	 • Create a Treasure token.
//	 • Put a menace counter on a creature you control.
//	 • That creature's controller mills cards equal to its power."
//
// Every mode targets a creature and destroys it first (CR 608.2c, printed
// order); the bullet follows. The mill reads the creature's power as it was
// (last-known information, CR 608.2h) and its controller as it was, so a
// creature that did not die still counts and one that left is read from its
// record. The menace counter is a keyword counter (CR 122.1b) on a creature
// of your choice at resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f64eff0e-dda7-457b-adf9-aa4e537f4f58",
		Name:         "Common Black Removal",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: ChooseOne(
			ModeDoing("Destroy target creature, then create a Food token.", gapB29CreatureTarget(),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					if err := gapB29DestroyModeTarget(ctx, occ); err != nil {
						return err
					}
					return CreateToken{Template: FoodToken(), N: 1}.Apply(ctx)
				}),
			ModeDoing("Destroy target creature, then create a Treasure token.", gapB29CreatureTarget(),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					if err := gapB29DestroyModeTarget(ctx, occ); err != nil {
						return err
					}
					return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
				}),
			ModeDoing("Destroy target creature, then put a menace counter on a creature you control.", gapB29CreatureTarget(),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					if err := gapB29DestroyModeTarget(ctx, occ); err != nil {
						return err
					}
					return ChoosePermanents{
						Question:   "Common Black Removal — choose a creature you control to put a menace counter on",
						Candidates: humiliateCreatures,
						Keyed:      commonBlackRemovalMenace,
					}.Apply(ctx)
				}),
			ModeDoing("Destroy target creature, then its controller mills cards equal to its power.", gapB29CreatureTarget(),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					card, found := ctx.Game.LookupCardForEffect(t.ID)
					if !found {
						return nil
					}
					controller := card.Controller
					if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					return MillCards{Player: controller, N: departedCreaturePower(ctx.Game, t.ID)}.Apply(ctx)
				}),
		),
	})
}

func gapB29CreatureTarget() *game.TargetSpec { return TargetCreature("target creature") }

// gapB29DestroyModeTarget is each bullet's first sentence: the mode's
// destroy, which Titan of Industry's first bullet already is.
func gapB29DestroyModeTarget(ctx *Context, occ int) error {
	return titanOfIndustryDestroy(nil, ctx, occ)
}

// commonBlackRemovalMenace is the chosen creature taking its menace counter.
// An on-disk key: never renamed, never reused.
var commonBlackRemovalMenace = RegisterPermanentsPick("prompt-run/common-black-removal-menace", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		cards := picked.Cards()
		if len(cards) != 1 {
			return nil
		}
		return AddCounter{Target: cards[0], Kind: game.CounterMenace, N: 1}.Apply(ctx)
	},
})

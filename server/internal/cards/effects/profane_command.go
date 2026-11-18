package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Profane Command — Sorcery {X}{B}{B}:
//
//	"Choose two —
//	 • Target player loses X life.
//	 • Return target creature card with mana value X or less from your
//	   graveyard to the battlefield.
//	 • Target creature gets -X/-X until end of turn.
//	 • Up to X target creatures gain fear until end of turn. (They can't
//	   be blocked except by artifact creatures and/or black creatures.)"
//
// Kozilek's Command's shape: choose two of four X-scaled bullets, each with
// its own target group, resolved in printed order (CR 608.2c). The
// reanimation bullet binds the card's mana value to the announced X
// (WithManaValueAtMostX); the last bullet's count is capped at the
// announced X (UpToX), so X=0 names no creature. A bullet whose target has
// gone does nothing and the other still resolves.
//
// No simplification.
func init() {
	fearTargets := TargetCreature("up to X target creatures")
	fearTargets.CountFromX = true
	fearTargets.UpToX = true
	Register(Spec{
		OracleID:     "6f9ecfc4-85bd-4171-9031-dbe0e39e0135",
		Name:         "Profane Command",
		XMatters:     true,
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			ModeDoing("Target player loses X life.",
				TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -ctx.X())
				}),
			ModeDoing("Return target creature card with mana value X or less from your graveyard to the battlefield.",
				TargetCardInGraveyard("target creature card with mana value X or less from your graveyard", YouOwn(), Creature()).WithManaValueAtMostX(),
				rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
					return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(ctx)
				})),
			ModeWithPurpose(ModeDoing("Target creature gets -X/-X until end of turn.",
				TargetCreature("target creature"),
				rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
					return BoostUntilEOT{Target: id, Power: -ctx.X(), Toughness: -ctx.X(), Label: "Profane Command"}.Apply(ctx)
				})), ForTargets(HarmUntilEOT(0, 0, 0))),
			ModeWithPurpose(ModeDoing("Up to X target creatures gain fear until end of turn.",
				fearTargets,
				func(_ *game.StackItem, ctx *Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"fear"}, Label: "Profane Command"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}), ForTargets(BenefitUntilEOT(0, 0, 0, "fear"))),
		),
	})
}

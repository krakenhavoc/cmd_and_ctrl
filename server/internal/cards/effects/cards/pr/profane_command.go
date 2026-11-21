package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	fearTargets := effects.TargetCreature("up to X target creatures")
	fearTargets.CountFromX = true
	fearTargets.UpToX = true
	effects.Register(effects.Spec{
		OracleID:     "6f9ecfc4-85bd-4171-9031-dbe0e39e0135",
		Name:         "Profane Command",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeDoing("Target player loses X life.",
				effects.TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -ctx.X())
				}),
			effects.ModeDoing("Return target creature card with mana value X or less from your graveyard to the battlefield.",
				effects.TargetCardInGraveyard("target creature card with mana value X or less from your graveyard", effects.YouOwn(), effects.Creature()).WithManaValueAtMostX(),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(ctx)
				})),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets -X/-X until end of turn.",
				effects.TargetCreature("target creature"),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.BoostUntilEOT{Target: id, Power: -ctx.X(), Toughness: -ctx.X(), Label: "Profane Command"}.Apply(ctx)
				})), effects.ForTargets(effects.HarmUntilEOT(0, 0, 0))),
			effects.ModeWithPurpose(effects.ModeDoing("Up to X target creatures gain fear until end of turn.",
				fearTargets,
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"fear"}, Label: "Profane Command"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "fear"))),
		),
	})
}

package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silverquill Command — Sorcery {2}{W}{B} (Strixhaven: School of Mages,
// #5502):
//
//	"Choose two —
//	 • Target creature gets +3/+3 and gains flying until end of turn.
//	 • Return target creature card with mana value 2 or less from your
//	   graveyard to the battlefield.
//	 • Target player draws a card and loses 1 life.
//	 • Target opponent sacrifices a creature of their choice."
//
// Choose two different bullets (CR 700.2d), each with its own targets, carried
// out in printed order (CR 608.2c) and re-checked one by one as the spell
// resolves (CR 608.2b): a bullet whose target is gone does nothing while the
// other still resolves. The draw comes before the life loss; the sacrifice is
// the opponent's own choice among the creatures they control (a creature the
// first bullet pumped included), made as the bullet resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6075b999-75c5-41c8-898d-ee612ba79a92",
		Name:         "Silverquill Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeWithPurpose(effects.ModeSteps("Target creature gets +3/+3 and gains flying until end of turn.",
				effects.TargetCreature("target creature"),
				effects.Instruction(func(ctx *effects.Context) error {
					t, ok := effects.ModeTarget(ctx, ctx.Occurrence())
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 3, Label: "Silverquill Command — +3/+3"}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					t, ok := effects.ModeTarget(ctx, ctx.Occurrence())
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: "Silverquill Command — flying"}.Apply(ctx)
				})),
				effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "flying"))),
			effects.ModeDoing("Return target creature card with mana value 2 or less from your graveyard to the battlefield.",
				effects.TargetCardInGraveyard("target creature card with mana value 2 or less from your graveyard", effects.YouOwn(), effects.Creature(), effects.ManaValueLE(2)),
				effects.ModeReturnTargetToTheBattlefield),
			effects.ModeWithPurpose(effects.ModeSteps("Target player draws a card and loses 1 life.",
				effects.TargetPlayer("target player"),
				effects.Instruction(func(ctx *effects.Context) error {
					t, ok := effects.ModeTarget(ctx, ctx.Occurrence())
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.DrawCards{Player: t.ID, N: 1}.Apply(ctx)
				}),
				effects.ModeTargetPlayerLosesLife(1)),
				effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 1, LifeLoss: 1})),
			effects.ModeDoing("Target opponent sacrifices a creature of their choice.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					ctx.Game.PlayerSacrificesForEffect(ctx.Source(), t.ID, effects.SacrificeSpec("a creature", effects.Creature()),
						"Silverquill Command — sacrifice a creature")
					return nil
				}),
		),
	})
}

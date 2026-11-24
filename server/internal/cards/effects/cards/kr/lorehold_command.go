package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lorehold Command — Instant {3}{R}{W} (Strixhaven: School of Mages, #5502):
//
//	"Choose two —
//	 • Create a 3/2 red and white Spirit creature token.
//	 • Creatures you control get +1/+0 and gain indestructible and haste until
//	   end of turn.
//	 • Lorehold Command deals 3 damage to any target. Target player gains 3
//	   life.
//	 • Sacrifice a permanent, then draw two cards."
//
// Choose two different bullets (CR 700.2d), carried out in printed order
// (CR 608.2c). The pump bullet fixes its set as it resolves (CR 611.2c): the
// creatures you control then, a Spirit made by the first bullet included. The
// damage bullet has two targets of its own, any target and a player, each
// re-checked at resolution; the life gain is separate from the damage, so it
// happens even when the damage target became illegal, and the other way
// round. The last bullet's draw follows the sacrifice, which is not optional:
// with no permanent to sacrifice (rare) the draw still happens.
//
// No simplification.
func init() {
	you := effects.And(effects.Creature(), effects.YouControl())
	effects.Register(effects.Spec{
		OracleID:     "107becb1-f8a3-4777-b39f-a0456c2d580e",
		Name:         "Lorehold Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeWithPurpose(effects.ModeDoing("Create a 3/2 red and white Spirit creature token.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("3/2 red and white Spirit"), N: 1}.Apply(ctx)
				}), game.Purpose{Tokens: 1}),
			effects.ModeSteps("Creatures you control get +1/+0 and gain indestructible and haste until end of turn.", nil,
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.BoostUntilEOT{Match: you, Power: 1, Label: "Lorehold Command — +1/+0"}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.GrantKeywordUntilEOT{Match: you, Keywords: []string{"indestructible", "haste"},
						Label: "Lorehold Command — indestructible and haste"}.Apply(ctx)
				})),
			effects.ModeWithPurpose(effects.ModeSteps("Lorehold Command deals 3 damage to any target. Target player gains 3 life.",
				effects.Clauses(effects.TargetAny(), effects.TargetPlayer("target player")),
				effects.Instruction(func(ctx *effects.Context) error {
					ts := ctx.ModeTargets(ctx.Occurrence())
					if len(ts) == 0 || !ctx.IsTargetLegal(ts[0]) {
						return nil
					}
					return effects.DealDamage{Source: ctx.Source(), Target: ts[0].ID, Amount: 3}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					ts := ctx.ModeTargets(ctx.Occurrence())
					if len(ts) < 2 || ts[1].Kind != game.TargetPlayer || !ctx.IsTargetLegal(ts[1]) {
						return nil
					}
					return effects.GainLife{Player: ts[1].ID, Amount: 3}.Apply(ctx)
				})),
				effects.ForTargets(effects.DamageToTarget(0, 3), game.TargetPurpose{Slot: 1, LifeGain: 3})),
			effects.ModeWithPurpose(effects.ModeSteps("Sacrifice a permanent, then draw two cards.", nil,
				effects.Instruction(func(ctx *effects.Context) error {
					ctx.Game.PlayerSacrificesForEffect(ctx.Source(), ctx.Controller(), effects.SacrificeSpec("a permanent", effects.Permanent()),
						"Lorehold Command — sacrifice a permanent")
					return nil
				}),
				effects.Instruction(func(ctx *effects.Context) error { return effects.DrawCards{N: 2}.Apply(ctx) })),
				game.Purpose{Draws: 2}),
		),
	})
}

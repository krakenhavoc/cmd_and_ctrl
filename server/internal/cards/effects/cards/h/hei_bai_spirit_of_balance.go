package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hei Bai, Spirit of Balance — Legendary Creature — Bear Spirit
// {2}{W/B}{W/B}, 3/3 (Avatar: The Last Airbender, #4459):
//
//	"Whenever Hei Bai enters or attacks, you may sacrifice another
//	 creature or artifact. If you do, put two +1/+1 counters on Hei Bai.
//	 When Hei Bai leaves the battlefield, put its counters on target
//	 creature you control."
//
// The first ability is one trigger with two conditions (Sun Titan's
// shape), its sacrifice chosen as it resolves (maySacrificeOne); the
// counters go on Hei Bai only if he is still the same object (CR 400.7).
//
// The leaves trigger reads Hei Bai's counters as he last existed on the
// battlefield (CR 608.2h) and puts the same number of each kind on the
// target (b14PutCounters) — -1/-1 counters and every other kind too, not
// just +1/+1 counters (the rulings). He is not a creature you control by
// then, so he cannot be the target. With no counters it does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f709660-bb26-409f-9650-e5228ac8db7e",
		Name:         "Hei Bai, Spirit of Balance",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Hei Bai — you may sacrifice another creature or artifact to put two +1/+1 counters on Hei Bai",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Hei Bai — sacrifice another creature or artifact to put two +1/+1 counters on Hei Bai?",
						effects.Or(effects.Creature(), effects.Artifact()), true, heiBaiSacrificed)
				}),
			effects.Targeting(effects.WhenThisLeaves("Hei Bai — put its counters on target creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					left, ok := ctx.TriggeringPermanent()
					if !ok || len(left.Counters) == 0 {
						return nil
					}
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.B14PutCounters(ctx, t.ID, left.Counters)
						}
					}
					return nil
				}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}

var heiBaiSacrificed = effects.RegisterSacrificedOne("prompt-run/hei-bai-two-counters", func(ctx *effects.Context, _ uuid.UUID) error {
	if effects.SourceIsNewObject(ctx.Game, ctx.Item) || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
})

package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elder Cathar — {2}{W} Creature — Human Soldier 2/2 (Shadows over Innistrad
// via Duel Decks: Blessed vs. Cursed, S82 OLD b294):
//
//	"When this creature dies, put a +1/+1 counter on target creature you
//	 control. If that creature is a Human, put two +1/+1 counters on it
//	 instead."
//
// One trigger aimed at a creature you control as it goes on the stack (CR
// 603.3d). On resolution the target is checked again (CR 608.2b) and its
// effective subtypes decide between one counter and two (CR 205.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba9ac0a9-735d-4bca-9d24-586aca963872",
		Name:         "Elder Cathar",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Elder Cathar — put a +1/+1 counter on target creature you control (two if it's a Human)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
						return nil
					}
					n := 1
					if c, ok := g.LookupCardForEffect(item.Targets[0].ID); ok && c.HasSubtype(ctx.TextSubtype("Human")) {
						n = 2
					}
					return effects.AddCounter{Target: item.Targets[0].ID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
				}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}

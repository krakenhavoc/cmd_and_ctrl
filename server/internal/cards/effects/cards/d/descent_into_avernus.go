package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Descent into Avernus — Enchantment {2}{R} (EDHREC top 1,000, #3654):
//
//	"At the beginning of your upkeep, put two descent counters on this
//	 enchantment. Then each player creates X Treasure tokens and this
//	 enchantment deals X damage to each player, where X is the number of
//	 descent counters on this enchantment."
//
// X is read after the two counters have landed (the "then"), from the
// enchantment's own counters at that point, so a doubler that changed
// how many were placed is counted. Every player makes their Treasures
// first, then every player takes the damage, the enchantment being the
// source of it. An enchantment that left before the trigger resolved
// does nothing.
//
// No simplification.
func init() {
	const label = "Descent into Avernus — put two descent counters, then each player creates X Treasures and takes X damage"
	effects.Register(effects.Spec{
		OracleID:     "1313da2f-92bb-4862-9153-373972cc1520",
		Name:         "Descent into Avernus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventStepBegan, effects.StepBegan(game.StepUpkeep, true), label, effects.Do(
				// The Treasures wait for a prompt the counters ask (CR 608.2c, #4155).
				effects.Instruction(func(ctx *effects.Context) error {
					src := ctx.Item.SourceCardID
					if !effects.OnBattlefield(ctx.Game, src) {
						return nil
					}
					return effects.AddCounter{Target: src, Kind: "descent", N: 2}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					g, src := ctx.Game, ctx.Item.SourceCardID
					if !effects.OnBattlefield(g, src) {
						return nil
					}
					c, ok := g.LookupCardForEffect(src)
					if !ok || c.Counters["descent"] <= 0 {
						return nil
					}
					x := c.Counters["descent"]
					players := effects.SeatsFromActive(g)
					for _, p := range players {
						if err := (effects.CreateToken{Controller: p, Template: effects.TreasureToken(), N: x}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return g.DamageInstanceForEffect(func() error {
						for _, p := range players {
							if err := (effects.DealDamage{Source: src, Target: p, Amount: x}).Apply(ctx.AsGroupMember()); err != nil {
								return err
							}
						}
						return nil
					})
				}),
			)),
		},
	})
}

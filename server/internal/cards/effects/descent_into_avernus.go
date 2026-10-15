package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "1313da2f-92bb-4862-9153-373972cc1520",
		Name:         "Descent into Avernus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, StepBegan(game.StepUpkeep, true), label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				src := item.SourceCardID
				if !onBattlefield(g, src) {
					return nil
				}
				if err := (AddCounter{Target: src, Kind: "descent", N: 2}).Apply(ctx); err != nil {
					return err
				}
				c, ok := g.LookupCardForEffect(src)
				if !ok || c.Counters["descent"] <= 0 {
					return nil
				}
				x := c.Counters["descent"]
				players := seatsFromActive(g)
				for _, p := range players {
					if err := (CreateToken{Controller: p, Template: TreasureToken(), N: x}).Apply(ctx.asGroupMember()); err != nil {
						return err
					}
				}
				return g.DamageInstanceForEffect(func() error {
					for _, p := range players {
						if err := (DealDamage{Source: src, Target: p, Amount: x}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				})
			}),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Compel Brutality — Instant {1}{G}:
//
//	"Choose one —
//	 • Target creature you control deals damage equal to its power to
//	   target creature or planeswalker an opponent controls.
//	 • Target planeswalker you control deals damage equal to its loyalty
//	   to target creature or planeswalker an opponent controls."
//
// Bite Down's one-sided bite in two flavours. Each bullet is a
// two-clause group (slot 0 deals, slot 1 takes) and the damage is read
// as the spell resolves, from the dealer's current power or its loyalty
// counters. The dealer is the damage's source, so deathtouch or
// lifelink on it applies. If either target is illegal on resolution
// no damage is dealt (CR 608.2b).
//
// No simplifications.
func init() {
	victim := TargetPermanent("target creature or planeswalker an opponent controls",
		Or(Creature(), Planeswalker()), OpponentControls())
	bite := func(amount func(c game.Card) int) func(*game.StackItem, *Context, int) error {
		return func(_ *game.StackItem, ctx *Context, occ int) error {
			dealer, ok := ModeClauseTarget(ctx, occ, 0)
			if !ok || dealer.Kind != game.TargetCard {
				return nil
			}
			taker, ok := ModeClauseTarget(ctx, occ, 1)
			if !ok || taker.Kind != game.TargetCard {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(dealer.ID)
			if !ok {
				return nil
			}
			return DealDamage{Source: dealer.ID, Target: taker.ID, Amount: amount(c)}.Apply(ctx)
		}
	}
	Register(Spec{
		OracleID:     "3a9059ac-c595-4e71-b704-b04d2d47db70",
		Name:         "Compel Brutality",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Target creature you control deals damage equal to its power to target creature or planeswalker an opponent controls.",
				Clauses(TargetCreature("target creature you control", YouControl()), victim),
				bite(func(c game.Card) int { return c.CurrentPower() })),
			ModeDoing("Target planeswalker you control deals damage equal to its loyalty to target creature or planeswalker an opponent controls.",
				Clauses(TargetPermanent("target planeswalker you control", Planeswalker(), YouControl()), victim),
				bite(func(c game.Card) int { return c.Counters[game.CounterLoyalty] })),
		),
	})
}

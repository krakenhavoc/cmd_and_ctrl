package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bottle-Cap Blast — Instant {4}{R} (Fallout):
//
//	"Improvise (Your artifacts can help cast this spell. Each artifact you
//	 tap after you're done activating mana abilities pays for {1}.)
//	 Bottle-Cap Blast deals 5 damage to any target. If excess damage was
//	 dealt to a permanent this way, create that many tapped Treasure
//	 tokens."
//
// Improvise is Spec.TapCost (Battle at the Bridge). The damage is Unleash
// the Inferno's: excess is what landed beyond lethal damage for a creature,
// beyond loyalty for a planeswalker, beyond defense for a battle (CR
// 120.10), read after the damage so a prevention shield lowers it. A player
// has no excess, so damage to a player makes nothing. The Treasures enter
// tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e9c78601-4bec-4d16-8fd9-88c6053bd9c4",
		Name:         "Bottle-Cap Blast",
		Completeness: CompletenessFull,
		TapCost:      Improvise(),
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			t := item.Targets[0]
			if t.Kind != game.TargetCard {
				return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 5}.Apply(ctx)
			}
			excess, err := s82MomB12DealDamageWithExcess(ctx, t.ID, 5)
			if err != nil || excess <= 0 {
				return err
			}
			return CreateTokenAdvanced{Spec: Token(TreasureToken()).EntersTapped(), N: excess}.Apply(ctx)
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Contest of Claws — Sorcery {1}{G}:
//
//	"Target creature you control deals damage equal to its power to
//	 another target creature. If excess damage was dealt this way,
//	 discover X, where X is that excess damage."
//
// Bite Down's shape with Hell to Pay's arithmetic. Excess damage is
// CR 120.4a's "more than lethal": what landed, less the damage the
// victim still needed to die (toughness less damage already marked,
// floored at zero), read after the damage so a prevention shield
// reduces the excess with the damage. No excess, or a target gone,
// discovers nothing. Discover is ADR 0099's (game/discover.go).
func init() {
	Register(Spec{
		OracleID:     "e30d5114-40d7-4dc5-8067-9c5e9625e097",
		Name:         "Contest of Claws",
		Completeness: CompletenessFull,
		Discovers:    true,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			Distinct(TargetCreature("another target creature")),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			biter, ok := ctx.ClauseTarget(0)
			if !ok || biter.Kind != game.TargetCard {
				return nil
			}
			victim, ok := ctx.ClauseTarget(1)
			if !ok || victim.Kind != game.TargetCard {
				return nil
			}
			attacker, ok := ctx.Game.LookupCardForEffect(biter.ID)
			if !ok {
				return nil
			}
			target, ok := ctx.Game.LookupCardForEffect(victim.ID)
			if !ok {
				return nil
			}
			lethal := target.CurrentToughness() - target.DamageMarked
			if lethal < 0 {
				lethal = 0
			}
			cursor := b25LastEventSeq(ctx.Game)
			if err := (DealDamage{Source: biter.ID, Target: victim.ID, Amount: attacker.CurrentPower()}).Apply(ctx); err != nil {
				return err
			}
			excess := b27DamageDealtToAfter(ctx.Game, biter.ID, victim.ID, cursor) - lethal
			if excess <= 0 {
				return nil
			}
			return Discover{N: excess}.Apply(ctx)
		},
	})
}

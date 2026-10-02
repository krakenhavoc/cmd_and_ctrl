package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// effects_damage_each_creature_and_player.go — "This creature deals N
// damage to each creature with[out] flying and each player", as an
// activated ability's whole body. The Mongers' shape (Squallmonger,
// Warmonger) and Ifh-Bíff Efreet's, first written for the any-player
// cards (ADR 0106 §1, #1793).
//
// The source deals the damage (CR 120.3): ctx.Source(), read as it
// last existed if it has left (CR 113.7a, 608.2h). The creatures are
// the ones matching when the ability resolves, read post-layer, so a
// creature that was granted flying is hit by Squallmonger and spared by
// Warmonger. Every player still in the game is hit, the activator and
// the source's controller included. Damage, not destruction: lethal
// damage kills at the next state-based check, and prevention works as
// printed.
//
// Append-only: add a builder, never change what one means.

// thisDealsDamageToEachCreatureMatchingAndEachPlayer is the body:
// `n` damage to each creature matching `match`, then to each player.
func thisDealsDamageToEachCreatureMatchingAndEachPlayer(match CardPredicate, n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		return ctx.Game.DamageInstanceForEffect(func() error {
			if err := damageEachMatching(ctx, And(Creature(), match), n); err != nil {
				return err
			}
			for _, p := range g.Seats {
				if p == nil || p.Eliminated {
					continue
				}
				if err := (DealDamage{Source: ctx.Source(), Target: p.ID, Amount: n}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	}
}

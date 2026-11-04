package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wave of Reckoning — Sorcery {4}{W} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Each creature deals damage to itself equal to its power."
//
// The damage is dealt BY each creature (the DealDamage source is the creature
// itself), so deathtouch, lifelink, infect and a damage doubler or prevention
// shield on it apply as in paper. Every creature's power is read as the spell
// resolves, once (CR 608.2h), and all the damage is dealt as one damage
// instance, so a creature's death does not change another creature's power
// first. A creature with power 0 or less deals no damage (CR 120.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24f7143c-d43a-4120-baa0-0235b15d8186",
		Name:         "Wave of Reckoning",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			type hit struct {
				id    game.Card
				power int
			}
			var hits []hit
			for _, id := range ctx.CreatureIDs() {
				c, ok := ctx.Game.LookupCardForEffect(id)
				if !ok {
					continue
				}
				if p := c.CurrentPower(); p > 0 {
					hits = append(hits, hit{id: c, power: p})
				}
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, h := range hits {
					if err := (DealDamage{Source: h.id.InstanceID, Target: h.id.InstanceID, Amount: h.power}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}

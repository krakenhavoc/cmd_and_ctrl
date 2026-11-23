package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orbital Plunge — Sorcery {3}{R} (Edge of Eternities, #4461):
//
//	"Orbital Plunge deals 6 damage to target creature. If excess damage
//	 was dealt this way, create a Lander token."
//
// Hell to Pay's damage (b27DealDamageWithExcess): the creature's lethal
// requirement, toughness less damage already marked, is read before the
// damage lands and what landed after, so a prevention shield lowers the
// excess along with the damage. Any excess at all (not a count of it)
// makes one Lander (b16LanderToken). A target that is gone fizzles the
// spell: no damage and no Lander.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f10b164-6d36-4935-92e7-53be73ac4c8a",
		Name:         "Orbital Plunge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				excess, err := effects.B27DealDamageWithExcess(ctx, t.ID, 6)
				if err != nil || excess <= 0 {
					return err
				}
				return effects.CreateToken{Controller: item.Controller, Template: effects.B16LanderToken(), N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}

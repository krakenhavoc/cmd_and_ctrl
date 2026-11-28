package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unerring Sling — Artifact {3} (Mirage, #5633):
//
//	"{3}, {T}, Tap an untapped creature you control: This artifact deals damage
//	 equal to the tapped creature's power to target attacking or blocking
//	 creature with flying."
//
// The tapped creature's power is read as the ability resolves, or as it last
// existed if it has left (CR 608.2h): Context.TappedPower. The creature is
// tapped as a cost, not with the {T} symbol, so one that arrived this turn can
// pay it (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10a52a73-157f-475e-975a-eb00c80573d4",
		Name:         "Unerring Sling",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{3}, {T}, Tap an untapped creature you control: This artifact deals damage equal to the tapped creature's power to target attacking or blocking creature with flying.",
			Cost: effects.Plus(effects.ManaCost("{3}"), effects.TapCost(),
				effects.TapAnUntapped("an untapped creature you control", effects.Creature())),
			Targets: effects.TargetCreature("target attacking or blocking creature with flying", effects.AttackingOrBlocking(), effects.HasKeyword("flying")),
			Effect: effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int {
				p, _ := ctx.TappedPower()
				if p < 0 {
					return 0
				}
				return p
			}),
		}},
	})
}

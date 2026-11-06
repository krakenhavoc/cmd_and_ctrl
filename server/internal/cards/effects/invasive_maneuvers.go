package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasive Maneuvers — Instant {1}{R} (Edge of Eternities, #4461):
//
//	"Invasive Maneuvers deals 3 damage to target creature. It deals 5
//	 damage instead if you control a Spacecraft."
//
// Plasma Bolt's shape with a creature target and the Spacecraft check in
// place of void. Whether you control a Spacecraft is read once, as the
// spell resolves (CR 608.2h), as the printed ruling says; a Spacecraft
// below its station threshold counts, since "Spacecraft" is the artifact
// subtype.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c376349-1a50-4215-8a40-f702da81eb7b",
		Name:         "Invasive Maneuvers",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			n := 3
			if ControlsA("Spacecraft")(ctx.Game, ctx.Controller()) {
				n = 5
			}
			return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: n}.Apply(ctx)
		},
	})
}

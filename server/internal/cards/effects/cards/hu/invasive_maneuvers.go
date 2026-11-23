package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4c376349-1a50-4215-8a40-f702da81eb7b",
		Name:         "Invasive Maneuvers",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			n := 3
			if effects.ControlsA("Spacecraft")(ctx.Game, ctx.Controller()) {
				n = 5
			}
			return effects.DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: n}.Apply(ctx)
		},
	})
}

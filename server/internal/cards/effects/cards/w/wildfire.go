package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wildfire — Sorcery {4}{R}{R} (Urza's Saga, S82 OLD b723):
//
//	Each player sacrifices four lands of their choice. Wildfire deals 4 damage to each creature.
//
// Destructive Force's shape: every living player is asked, in APNAP order, for four lands (fewer when they control fewer), the choices are sacrificed together (CR 101.4, 701.21a), and the damage waits for the sacrifices and is dealt to every creature as one instance, the spell being the source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0da68ca2-07fd-4359-95a9-fb2f815e6ed0",
		Name:         "Wildfire",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.PlayersSacrificeThen(effects.S82OldB475SeatsOwing(ctx.Game, 4),
				effects.SacrificeSpec("a land", effects.Land()), "Wildfire — sacrifice four lands",
				effects.S82OldB723WildfireDamage, ctx.RunData(nil))
		},
	})
}

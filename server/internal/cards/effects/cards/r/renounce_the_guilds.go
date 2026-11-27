package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Renounce the Guilds — Instant {1}{W} (Dragon's Maze, S82 OLD batch 403, #5564):
//
//	"Each player sacrifices a multicolored permanent of their choice."
//
// Barter in Blood's edict over every player, the caster included, one choice each
// in APNAP order, sacrificed together (CR 701.21a, 101.4): each player chooses a
// permanent of two or more colours they control, read through the layers; a
// player with exactly one is not asked and one with none sacrifices nothing.
// Nothing is targeted, so hexproof and indestructible do not save a permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c854f9e8-6a1c-4515-9912-3847978bf5d5",
		Name:         "Renounce the Guilds",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(effects.EachPlayerSacrifices{
			Match: effects.Multicolored(),
			Label: "a multicolored permanent",
		}),
	})
}

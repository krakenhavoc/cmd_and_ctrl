package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Psychotic Haze — Instant {2}{B}{B} (Torment, #5612):
//
//	"Psychotic Haze deals 1 damage to each creature and each player.
//	 Madness {1}{B} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// Inferno's body at 1 (b23DamageEachCreatureAndEachPlayer: one damage
// event, the caster and their creatures included) with madness, which
// the engine grows from the one string (CR 702.35a, #657).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b6aa5e2-b539-46cd-804b-cd9e996a9456",
		Name:         "Psychotic Haze",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
		Madness:      "{1}{B}",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b23DamageEachCreatureAndEachPlayer(ctx, 1)
		},
	})
}

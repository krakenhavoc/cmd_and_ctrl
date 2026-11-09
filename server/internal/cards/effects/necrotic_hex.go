package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necrotic Hex — Sorcery {6}{B} (Foundations Commander):
//
//	"Each player sacrifices six creatures of their choice. You create six
//	 tapped 2/2 black Zombie creature tokens."
//
// Each living player, in turn order from the active player, names six
// creatures (or all they have, CR 701.21a), each seeing the choices before
// theirs, and everything named is sacrificed together (printed ruling). The
// Zombies come after, whatever was sacrificed, and cannot be among the
// victims since they do not exist when the choices are made.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "a6e720a2-9819-40ec-a811-7812898b531a",
		Name:     "Necrotic Hex",
		Purpose: game.Purpose{
			Sweep:          game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true},
			Tokens:         6,
			CreatureTokens: 6,
		},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var seats []uuid.UUID
			for _, seat := range apnapLivingSeats(ctx.Game) {
				for i := 0; i < 6; i++ {
					seats = append(seats, seat)
				}
			}
			return ctx.PlayersSacrificeThen(seats,
				sacrificeSpec("a creature", Creature()), "Necrotic Hex — sacrifice six creatures",
				necroticHexZombies, ctx.RunData(nil))
		},
	})
}

// necroticHexZombies is the second sentence, once the sacrifices have
// landed: six tapped 2/2 black Zombies for the spell's controller.
var necroticHexZombies = RunThen("prompt-run/necrotic-hex-zombies", func(ctx *Context, _ game.RunDone) error {
	return CreateTokenAdvanced{Spec: Token(TokenCard("2/2 black Zombie")).EntersTapped(), N: 6}.Apply(ctx)
})

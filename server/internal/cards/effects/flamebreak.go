package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flamebreak — Sorcery {R}{R}{R}:
//
//	"Flamebreak deals 3 damage to each creature without flying and each
//	 player. Creatures dealt damage this way can't be regenerated this
//	 turn."
//
// One batch over the creatures without flying (read post-layer as it
// resolves) and every player still in the game, with "creatures dealt
// damage this way" read from the damage's continuation (ADR 0108 §2):
// a creature whose damage was all prevented can still be regenerated,
// and a player is never marked.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "9e608ace-2844-419b-9204-9054127557b2",
		Name:         "Flamebreak",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return damageEachMatchingAndEachPlayerThen(ctx, And(Creature(), WithoutKeyword("flying")), 3,
				CantBeRegeneratedIfDealtDamage(item))
		},
	})
}

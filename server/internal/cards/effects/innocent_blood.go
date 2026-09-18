package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Innocent Blood — Sorcery {B} (EDHREC rank 4044):
//
//	"Each player sacrifices a creature of their choice."
//
// A one-mana symmetrical edict. It is in the batch because it is the
// cheapest card in the catalog that makes every seat at the table
// choose at once, and because "of their choice" is the clause that
// makes it a sacrifice rather than a removal spell: nothing targets,
// so hexproof, shroud, protection and ward do not save anything, and
// each player picks their own worst creature rather than the caster
// picking their best.
//
// The caster is included — the printed text says "each player", not
// "each opponent" — so a mono-black deck plays it with an expendable
// token or with nothing on board at all. A player with no creature
// sacrifices nothing and is skipped; that is not a failure to resolve.
//
// One prompt per player, each offering only that player's own
// creatures, through the shared each-player-sacrifices primitive.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6791ec3c-c397-4087-8c8c-84d3797df415",
		Name:         "Innocent Blood",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return EachPlayerSacrifices{Match: Creature(), Label: "a creature"}.Apply(ctx)
		},
	})
}

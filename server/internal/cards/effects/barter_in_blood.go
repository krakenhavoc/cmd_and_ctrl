package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barter in Blood — Sorcery {2}{B}{B}:
//
//	"Each player sacrifices two creatures of their choice."
//
// Liliana, Dreadhorde General's −4 as a spell: each player, the caster
// included, makes one choice of two creatures in APNAP order, and the
// pair is sacrificed together (CR 701.21a). A player with one creature
// sacrifices it; a player with none is not asked. Nothing is targeted,
// so hexproof and indestructible do not save a creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9167998d-5cac-47d7-99f2-f38122f7b8e7",
		Name:         "Barter in Blood",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return jmpB1EachPlayerSacrificesN(2, "Barter in Blood — choose two creatures to sacrifice")(ctx.Game, item)
		},
	})
}

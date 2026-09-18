package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flame Rift — Sorcery {1}{R} (EDHREC rank 3994):
//
//	"Flame Rift deals 4 damage to each player."
//
// Two mana for sixteen damage spread across a four-player table, of
// which four points come back at the caster. It is in the batch as
// the smallest possible card that hits EVERY seat including its own
// controller — the corner a "deals damage to each opponent" card
// never exercises — and because the group-slug decks that play it
// (Mogis, Torbran) need the symmetry to actually be symmetric.
//
// Damage from a SPELL, not a creature, so it is noncombat damage with
// no colour of its own: Torbran's "red sources you control" bonus
// applies, a player's protection from red prevents it, and nothing
// about it is combat damage for the cards that care.
//
// An eliminated seat is skipped — a player who has left the game is
// no longer a player (CR 800.4) and takes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "28479aa8-d1cd-421a-8bbb-0594bb4dd410",
		Name:         "Flame Rift",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b38DamageEachPlayer(ctx, 4)
		},
	})
}

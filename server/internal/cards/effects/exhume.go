package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exhume — Sorcery {1}{B} (EDHREC rank 3955):
//
//	"Each player puts a creature card from their graveyard onto the
//	 battlefield."
//
// Two mana to reanimate, and the cheapest reanimation spell that
// works at all. The catch is printed on it: everyone gets one. A
// reanimator deck plays it anyway, because the game where an
// opponent's graveyard has a bigger threat than yours is a game you
// were losing already.
//
// It is in the batch as the SYMMETRIC corner of the reanimation
// family, which nothing else in the catalog covers. Every other
// reanimation card names a target; this one names none.
//
// # Nothing targets, and that is the card
//
// Each player chooses for themselves, from their own graveyard —
// so the caster cannot pick an opponent's worst creature, and an
// opponent cannot be stopped from picking their best. The choices
// are made as it resolves, one prompt per player, and a player with
// no creature card in their graveyard simply does nothing (CR
// 608.2): that is not a failure to resolve, and it does not stop
// anybody else.
//
// Each creature returns under its OWNER's control, which — because
// the card came out of that player's own graveyard — is that player.
// Nothing changes hands.
//
// The creatures arrive as ordinary entries, so enters-the-battlefield
// triggers fire and the usual entry replacements apply.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fbe61f74-1b3c-4e12-8758-7029872c9ff1",
		Name:         "Exhume",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b38EachPlayerReanimatesOne(ctx.Game, item,
				"Exhume — put a creature card from your graveyard onto the battlefield")
		},
	})
}

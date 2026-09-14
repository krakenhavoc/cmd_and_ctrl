package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aetherspouts — Instant {3}{U}{U} (EDHREC rank 2780):
//
//	"For each attacking creature, its owner puts it on their choice of
//	 the top or bottom of their library."
//
// Aetherize's big sibling: the attackers do not come back to hand,
// they go into the library, and each owner decides where. The owner's
// choice is asked as a scry — every attacker an owner owns is put on
// top of their library and the owner then scries that many, which is
// exactly the printed decision: each card to the bottom or kept on
// top, and the order of the ones kept (CR 401.4 hands the owner that
// order anyway). The scry shows nothing that was not public a moment
// earlier. Owners are asked in seat order; nobody without an attacker
// is asked.
//
// The set is every attacking creature, any controller — a creature
// the caster controls that was goaded into attacking goes too, as
// printed.
//
// One engine-wide gap, not this card's: a token put into a library
// (or a hand, by any bounce) is not swept out of it as CR 111.8
// asks — it sits there as a card. Reported on #388; Evacuation and
// Aetherize share it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48369aec-a991-4bef-8554-01c84302b063",
		Name:         "Aetherspouts",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b26TuckAttackersTopOrBottomByOwnersChoice(ctx)
		},
	})
}

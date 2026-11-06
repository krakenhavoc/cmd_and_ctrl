package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Calamitous Tide — Sorcery {4}{U}{U} (Bloomburrow, #4467):
//
//	"Return up to two target creatures to their owners' hands. Draw two
//	 cards, then discard a card."
//
// Horses of the Bruinen's return (the creatures go together, and a
// returned commander may stop for its owner's command-zone answer, CR
// 903.9), then Three Steps Ahead's loot in the return's keyed tail, the
// draw first so a drawn card is a legal discard. With no target chosen
// the spell still loots; with every chosen target illegal it does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d605d9d1-c01a-4ef6-89f9-06fae8cd57e6",
		Name:         "Calamitous Tide",
		Purpose:      game.Purpose{Draws: 2, Discards: 1},
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ctx.Game.BounceCardsToHandKeyedForEffect(legalTargetIDs(ctx), ctx.ItemTail(calamitousTideLoot))
		},
	})
}

// calamitousTideLoot is "Draw two cards, then discard a card." after the
// return, keyed (#4253): the key is an on-disk identity.
var calamitousTideLoot = CardTail("then/calamitous-tide-loot", func(ctx *Context, _ game.TailOutcome) error {
	return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 2, 1, "")
})

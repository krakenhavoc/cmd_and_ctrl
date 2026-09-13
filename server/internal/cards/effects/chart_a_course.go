package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chart a Course — Sorcery {1}{U} (EDHREC rank 1946):
//
//	"Draw two cards. Then discard a card unless you attacked this
//	 turn."
//
// Two cards for two mana, free after combat if you swung. "Attacked
// this turn" is a fact about the turn, not the board, and the
// engine keeps no flag for it — so it is read off the event log:
// an EventAttack by the caster more recent than the current turn's
// upkeep (b18AttackedThisTurn). The draw happens first, so the
// drawn cards are legal discards, and the discard is the player's
// own choice through the discard modal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "05878e49-93ad-4144-9c50-a0bb86126c2e",
		Name:         "Chart a Course",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
				return err
			}
			if b18AttackedThisTurn(ctx.Game, item.Controller) {
				return nil
			}
			ctx.Game.DiscardChoiceForEffect(item.Controller, 1)
			return nil
		},
	})
}

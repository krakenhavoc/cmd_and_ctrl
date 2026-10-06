package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Addle — Sorcery {1}{B}:
//
//	"Choose a color. Target player reveals their hand and you choose a
//	 card of that color from it. That player discards that card."
//
// A choice made before the revealed-hand pick (#2115, ADR 0116's
// 2026-10-05 amendment). The colour is chosen first, as the spell
// resolves and before the hand is seen (CR 608.2c). Its answer then
// raises the ordinary revealed-hand pick (ADR 0116) with "a card of
// that color" as the filter: the whole table sees the hand (CR
// 701.20a), and a hand with no card of that colour is revealed and
// nothing is discarded (CR 609.3). Colour is read as each card has it
// in the hand, so a multicoloured card is a card of each of its colours
// and a colourless one of none (CR 105.2).
//
// The pick is queued from the colour's answer while the resolution is
// still open, so state-based actions and triggers still wait for it
// (#1289).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3745aea4-4455-4859-9772-bcfd14e4067b",
		Name:         "Addle",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			victim := TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			chooser, source := ctx.Controller(), ctx.Source()
			ChooseColorThen(game.ColorForFilter, ctx.Game, chooser, source, "Addle — choose a color",
				func(g *game.Game, color string) error {
					next := NewContext(g, &game.StackItem{
						Kind: game.StackItemSpell, Controller: chooser, Owner: chooser, SourceCardID: source,
					})
					return ChooseFromRevealedHand{
						Player: victim,
						Filter: OfColor(color),
						Label:  game.ColorName(color) + " card",
					}.Apply(next)
				})
			return nil
		},
	})
}

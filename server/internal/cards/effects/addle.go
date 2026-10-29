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
			ChooseColorTail(game.ColorForFilter, ctx.Game, ctx.Controller(), ctx.Source(), "Addle — choose a color",
				withCarry(ctx.ItemTail(addleColorChosen), victim))
			return nil
		},
	})
}

// addleColorChosen is the rest of the spell once the colour (the tail's
// Word) is known: the target player, its one carried ID, reveals their hand
// and the caster chooses a card of that colour. Run against the asking spell.
// An on-disk key.
var addleColorChosen = CardTail("then/addle-color-chosen", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	color := o.Tail.Word
	return ChooseFromRevealedHand{
		Player: o.Tail.Carry[0],
		Filter: OfColor(color),
		Label:  game.ColorName(color) + " card",
	}.Apply(ctx)
})

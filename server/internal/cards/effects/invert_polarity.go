package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invert Polarity — Instant {U}{U}{R}:
//
//	"Choose target spell, then flip a coin. If you win the flip, gain
//	 control of that spell and you may choose new targets for it. If you
//	 lose the flip, counter that spell."
//
// Three existing shapes and ADR 0104's one new one. The flip is the
// called won/lost flip (CR 705.2, ADR 0054): the controller calls, and
// the engine draws won or lost when the call is answered. The branch
// runs from the flip's continuation, which captures only IDs, so an
// undo resolves it against the restored game.
//
// Won: GainControlOfSpell — a layer-2 effect pinned to the spell on the
// stack (CR 613.1b), after which the new controller is offered "you may
// choose new targets for it" (CR 115.7d). A permanent spell becomes a
// permanent under them, with its caster as its default controller
// (CR 110.2b).
//
// Lost: an ordinary counter, so a spell that can't be countered is
// neither stolen nor countered.
//
// The spell is a TARGET: if it has left the stack by resolution,
// Invert Polarity does nothing and nobody flips (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "69e12b1f-0fd9-43a9-b3db-fe08290442c6",
		Name:         "Invert Polarity",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			spell, controller, source, self := item.Targets[0].ID, item.Controller, item.SourceCardID, item.ID
			ctx.Game.FlipCoinForEffect(game.CoinFlipSpec{
				Flipper:  controller,
				Source:   source,
				Question: "Invert Polarity — call the coin flip",
				Then: func(g *game.Game, result game.CoinFlipResult) error {
					c := invertPolarityContext(g, self, controller, source)
					if len(result.Won) > 0 && result.Won[0] {
						return GainControlOfSpell{Spell: spell, ChooseNewTargets: true,
							Label: "Invert Polarity — gain control of that spell"}.Apply(c)
					}
					return CounterTarget{StackID: spell}.Apply(c)
				},
			})
			return nil
		},
	})
}

// invertPolarityContext rebuilds the resolving spell's context inside
// the flip's continuation from values, never from the item it began
// with.
func invertPolarityContext(g *game.Game, self, controller, source uuid.UUID) *Context {
	return NewContext(g, &game.StackItem{
		ID: self, Kind: game.StackItemSpell,
		Controller: controller, Owner: controller, SourceCardID: source,
	})
}

package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mystic Genesis — Instant {2}{G}{U}{U} (Gatecrash, #5565):
//
//	"Counter target spell. Create an X/X green Ooze creature token, where X
//	 is that spell's mana value."
//
// The spell's mana value is read before the counter, while it is still on
// the stack (an {X} in its cost counts as the X it was cast with, CR 202.3e),
// and the token is made whether or not the spell was actually countered; an
// X of 0 makes a 0/0 that is put into the graveyard at once. The token is
// created 0/0 and sized as it is made (Seed Guardian's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58940a7e-3827-43fa-8bf5-cc88dc295df0",
		Name:         "Mystic Genesis",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				x, _ := ctx.Game.LastKnownSpellManaValueForEffect(t.ID)
				if err := (effects.CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
					return err
				}
				tok := ctx.TextToken(effects.TokenCard("0/0 green Ooze"))
				tok.Power, tok.Toughness = x, x
				return ctx.Game.CreateTokenForEffect(ctx.Controller(), tok, 1)
			}
			return nil
		},
	})
}

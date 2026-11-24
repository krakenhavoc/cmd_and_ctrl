package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Starnheim Unleashed — Sorcery {2}{W}{W} (Kaldheim):
//
//	"Create a 4/4 white Angel Warrior creature token with flying and
//	 vigilance. If this spell was foretold, create X of those tokens instead.
//	 Foretell {X}{X}{W} (During your turn, you may pay {2} and exile this card
//	 from your hand face down. Cast it on a later turn for its foretell cost.)"
//
// X is the value chosen when the spell is cast for its foretell cost
// ({X}{X}{W}); a spell cast from hand has no X and makes one token. "Was
// foretold" is the spell's own status (CR 702.143c), true however the cast
// out of exile was paid for.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "495ebe60-f0e5-4f0f-9d99-b2c64c95dbf4",
		Name:           "Starnheim Unleashed",
		Completeness:   effects.CompletenessFull,
		XMatters:       true,
		SpecialActions: []game.SpecialAction{effects.Foretell("{X}{X}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 1
			if ctx.SpellWasForetold() {
				n = ctx.X()
			}
			if n <= 0 {
				return nil
			}
			return effects.CreateToken{Template: effects.TokenCard("4/4 white Angel Warrior with flying and vigilance"), N: n}.Apply(ctx)
		},
	})
}

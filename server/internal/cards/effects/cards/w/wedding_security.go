package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wedding Security — Creature — Vampire Soldier {3}{B}{B}, 4/4 (Innistrad:
// Crimson Vow, #5495):
//
//	"Whenever this creature attacks, you may sacrifice a Blood token. If
//	 you do, put a +1/+1 counter on this creature and draw a card."
//
// The sacrifice is asked as the trigger resolves (maySacrificeOne), over
// the Blood tokens you control then; the counter and the draw follow only
// if one was sacrificed. The counter goes on the Security as it is now
// (none if it has left), and the draw happens regardless.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86360908-6b8b-4c1c-9e10-4f3d389aaa3e",
		Name:         "Wedding Security",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Wedding Security — you may sacrifice a Blood token to put a +1/+1 counter on it and draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Wedding Security — sacrifice a Blood token? If you do, put a +1/+1 counter on it and draw a card.",
						effects.HasSubtype("Blood"), false, weddingSecurityPayoff)
				}),
		},
	})
}

var weddingSecurityPayoff = effects.RegisterSacrificedOne("prompt-run/wedding-security-counter-draw", func(ctx *effects.Context, _ uuid.UUID) error {
	if err := (effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
})

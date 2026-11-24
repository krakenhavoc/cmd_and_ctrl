package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Metallurgic Summonings — Enchantment {3}{U}{U} (Commander 2021, #5504):
//
//	"Whenever you cast an instant or sorcery spell, create an X/X
//	 colorless Construct artifact creature token, where X is that spell's
//	 mana value.
//	 {3}{U}{U}, Exile this enchantment: Return all instant and sorcery
//	 cards from your graveyard to your hand. Activate only if you control
//	 six or more artifacts."
//
// X is the spell's mana value as the trigger resolves, its announced X
// included while it is on the stack (CR 202.3e); the Construct is a
// colourless artifact creature whose size is set as it is created (an X of
// 0 makes a 0/0 that dies as a state-based action). The ability exiles the
// enchantment as a cost, is gated on the artifacts you control as it is
// activated (CR 602.5b), and returns every instant and sorcery card in
// your graveyard as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d5149c6-0e7e-438c-a554-5235c85e2426",
		Name:         "Metallurgic Summonings",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()), "Metallurgic Summonings — create an X/X Construct, X the spell's mana value",
				func(g *game.Game, item *game.StackItem) error {
					x := effects.TriggeringSpellManaValue(g, item)
					ctx := effects.NewContext(g, item)
					tok := ctx.TextToken(effects.TokenCard("0/0 colorless Construct artifact"))
					tok.Power, tok.Toughness = x, x
					return g.CreateTokenForEffect(ctx.Controller(), tok, 1)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "{3}{U}{U}, Exile this enchantment: Return all instant and sorcery cards from your graveyard to your hand. Activate only if you control six or more artifacts.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.Plus(effects.ManaCost("{3}{U}{U}"), effects.ExileThis()),
			Condition: effects.ControlsAtLeast(6, game.Card.IsArtifact),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ReturnWholeGraveyardToHand(effects.NewContext(g, item), func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() })
			},
		}},
	})
}

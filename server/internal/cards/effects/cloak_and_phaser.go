package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloak and Phaser — Sorcery {8}{U} (Star Trek):
//
//	Strange new worlds — This spell costs {X} less to cast, where X is the
//	number of differently named lands you control.
//	Create two 2/2 black Romulan creature tokens. Draw two cards, then
//	discard a card.
//
// Strange new worlds is an ability word (no rules meaning of its own); the
// reduction is a self cost modifier counting each land the caster controls
// once per name, and it reduces generic mana only (CR 601.2f), so {U} is
// always paid. The Romulans land before the draw, and the discard is the
// draw's continuation (lootOne's order).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "668a9b6b-731c-4127-9c88-1c6112ffd590",
		Name:         "Cloak and Phaser",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2, Draws: 2, Discards: 1},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(trekB5DifferentlyNamedLandsYouControl(),
				"Strange new worlds — This spell costs {X} less to cast, where X is the number of differently named lands you control."),
		},
		OnResolve: ResolveSteps(
			CreateToken{Template: TokenCard("2/2 black Romulan"), N: 2},
			Instruction(func(ctx *Context) error {
				return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 2, 1, "")
			}),
		),
	})
}

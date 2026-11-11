package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Welcome the Dead — Sorcery {3}{B} (Tarkir: Dragonstorm Commander):
//
//	"Draw two cards, then discard a card and you lose 2 life. Create X tapped
//	 2/2 black Zombie Druid creature tokens, where X is the number of cards
//	 that were put into your graveyard from your hand or library this turn.
//	 Flashback {5}{B} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// The loot is Calamitous Tide's (the draw first, so a drawn card is a legal
// discard, and the discard prompt is waited out before the next step, CR
// 608.2c). X is read after the discard and the life loss, off the per-turn
// tally of cards put into the controller's graveyard from a hand or a library
// (game.PlayerTurnTally.CardsToGraveyardFromHandOrLibrary): cards milled or
// discarded earlier this turn count, and so does the discard just made, and
// the spell itself (cast from hand, then on the stack) or a flashbacked copy
// from the graveyard does not. Flashback is the shared alternative cost.
//
// Seam: game.PlayerTurnTally.CardsToGraveyardFromHandOrLibrary.
func init() {
	Register(Spec{
		OracleID:         "affa0b6c-a9fd-49fc-b9ce-9a3486af8320",
		Name:             "Welcome the Dead",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Draws: 2, Discards: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{B}")},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 2, 1, "")
			}),
			Instruction(func(ctx *Context) error {
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
			}),
			Instruction(func(ctx *Context) error {
				x := ctx.Game.CardsToGraveyardFromHandOrLibraryThisTurn(ctx.Controller())
				if x <= 0 {
					return nil
				}
				tmpl := TokenCard("2/2 black Zombie Druid")
				tmpl.Tapped = true
				return CreateToken{Template: tmpl, N: x}.Apply(ctx)
			}),
		),
	})
}

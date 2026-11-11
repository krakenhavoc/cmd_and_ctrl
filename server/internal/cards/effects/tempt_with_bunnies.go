package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tempt with Bunnies — Sorcery {2}{W} (Bloomburrow Commander):
//
//	"Tempting Offer — Draw a card and create a 1/1 white Rabbit creature
//	 token. Then each opponent may draw a card and create a 1/1 white
//	 Rabbit creature token. For each opponent who does, you draw a card
//	 and you create a 1/1 white Rabbit creature token."
//
// Each opponent is asked in turn order; a yes gives that opponent the
// card and the Rabbit and gives the caster the same again.
func init() {
	tempt := newCmdr2S49Tempting("confirm/tempt-with-bunnies", "Tempt with Bunnies — draw a card and create a 1/1 Rabbit (the caster gets the same if you do)?", temptWithBunniesDo)
	Register(Spec{
		OracleID:     "6874be51-623d-4eaf-a67a-2fcbb9a8ceca",
		Name:         "Tempt with Bunnies",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Tokens: 1, CreatureTokens: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return tempt.resolve(ctx)
		},
	})
}

func temptWithBunniesDo(ctx *Context, player uuid.UUID) error {
	if err := (DrawCards{Player: player, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return CreateToken{Controller: player, Template: TokenCard("1/1 white Rabbit"), N: 1}.Apply(ctx)
}

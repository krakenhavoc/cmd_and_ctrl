package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slime Against Humanity — Sorcery {2}{G} (Murders at Karlov Manor,
// #4469):
//
//	"Create a 0/0 green Ooze creature token with trample. Put X +1/+1
//	 counters on it, where X is two plus the total number of cards you
//	 own in exile and in your graveyard that are Oozes or are named Slime
//	 Against Humanity.
//	 A deck can have any number of cards named Slime Against Humanity."
//
// The count is read as the spell resolves, so the spell itself (still on
// the stack) is not counted. Face-down exiled cards are not counted (the
// ruling). The token enters with the counters (CR 614.1c) rather than
// receiving them after, which is the same outcome, a 0/0 never being
// checked in between (CR 704.3). The deck-building line is not a game
// rule the engine enforces.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b53597f4-1a0f-4fa8-9c17-29178cdc4d2b",
		Name:         "Slime Against Humanity",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := 2 + slimeCount(ctx.Game, item.Controller)
			_, err := ctx.Game.CreateTokensForEffect(item.Controller,
				ctx.TextToken(TokenCard("0/0 green Ooze with trample")), 1,
				game.TokenEntryOptions{Counters: map[string]int{game.CounterPlusOne: x}})
			return err
		},
	})
}

// slimeCount is the cards `you` own, in exile (face up) and in their
// graveyard, that are Oozes or are named Slime Against Humanity.
func slimeCount(g *game.Game, you uuid.UUID) int {
	n := 0
	count := func(c game.Card) {
		if c.Owner == you && !c.FaceDown && (c.HasSubtypeWord("Ooze") || c.Name == "Slime Against Humanity") {
			n++
		}
	}
	if g.Exile != nil {
		for _, c := range g.Exile.Cards {
			count(c)
		}
	}
	if p := g.PlayerByIDForEffect(you); p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			count(c)
		}
	}
	return n
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// turn_control_cards.go — the shared half of the cards that say "you
// control target player during that player's next turn" (CR 723.1):
// Mindslaver, Worst Fears, Sorin Markov's -7, Construct a Cosmic Cube,
// Cruel Entertainment. The machinery is ADR 0167's turn-long control
// record (game.BeginTurnControlForEffect); these cards are the plain
// form of it, with no extra turn afterwards (TurnControl.NoExtraTurn,
// ADR 0167's 2026-10-16 amendment). Emrakul, the Promised End is the
// form that does owe one, and calls the engine directly.

// controlPlayerNextTurn schedules `controller`'s control of `controlled`
// during controlled's next turn, with no extra turn after it. The engine
// refuses a player who has left the game and a player "controlling"
// themselves, which is the printed outcome for a target that left in
// response (CR 608.2b) and for a player who targets themselves (nothing
// changes hands).
func controlPlayerNextTurn(g *game.Game, item *game.StackItem, controller, controlled uuid.UUID, sourceName string) {
	g.BeginTurnControlForEffect(game.TurnControl{
		Controller:  controller,
		Controlled:  controlled,
		Source:      item.SourceCardID,
		SourceName:  sourceName,
		NoExtraTurn: true,
	})
}

// firstLegalTargetPlayer is the first still-legal player in the item's
// targets (CR 608.2b: a target that left is skipped), or uuid.Nil.
func firstLegalTargetPlayer(ctx *Context) uuid.UUID {
	for _, ref := range ctx.LegalTargets() {
		if ref.Kind == game.TargetPlayer && ref.ID != uuid.Nil {
			return ref.ID
		}
	}
	return uuid.Nil
}

// controlTargetPlayerNextTurn is the effect of "You control target
// player during that player's next turn." for a card named sourceName:
// the item's controller controls the one player it targets.
func controlTargetPlayerNextTurn(sourceName string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if p := firstLegalTargetPlayer(ctx); p != uuid.Nil {
			controlPlayerNextTurn(g, item, item.Controller, p, sourceName)
		}
		return nil
	}
}

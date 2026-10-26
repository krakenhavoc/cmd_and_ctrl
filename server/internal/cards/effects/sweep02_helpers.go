package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep02_helpers.go — #4253's closure sweep, batch 02: the small shared
// shapes the batch's cards key their prompts with.

// itemChoice registers a keyed choose_cards continuation whose body runs
// against a Context rebuilt from the asking item (ChooseCardsPrompt.Asked,
// game.NewAskedItem). A prompt that carried no item runs against a stand-in
// with the chooser as the controller and the asking card as the source. Call
// it once, from a package-level var. The key is an on-disk identity
// ("choose-cards/<card>-<what>"), ledgered: never renamed, never reused.
func itemChoice(key string, body func(ctx *Context, r game.CardsChosen) error) game.ChooseCardsThen {
	return game.RegisterChooseCardsThen(key, func(g *game.Game, r game.CardsChosen) error {
		ctx, _ := askedContext(g, r.Asked, r.Chooser, r.Source, nil)
		return body(ctx, r)
	})
}

// ChoosePlayerKeyed is ChoosePlayer with the answer as a registered
// continuation instead of a closure (#4253), so a table waiting on the
// choice is a restore point. Register the body with ChoosePlayerThen and
// put what it reads beyond the player on Carry.
type ChoosePlayerKeyed struct {
	// Chooser is asked. Zero means the resolving effect's controller.
	Chooser uuid.UUID
	// Among is the pool the clause admits. Zero value is Players.
	Among PlayerPool
	// Except are seats the clause rules out.
	Except   []uuid.UUID
	Question string
	// Then is the registered answer; Carry the IDs it is handed.
	Then  game.OptionPickThen
	Carry []uuid.UUID
}

// ChoosePlayerThen registers the answer to a ChoosePlayerKeyed. The body is
// handed a Context rebuilt as an option pick's is (the asking effect's
// controller, the card that asked), the player chosen (uuid.Nil when
// nobody could be, or the prompt was dropped) and the carry the card put on
// the prompt. It runs on every outcome, the way ChoosePlayer.Then does.
func ChoosePlayerThen(key string, body func(ctx *Context, chosen uuid.UUID, carry []uuid.UUID) error) game.OptionPickThen {
	return OptionPickThen(key, func(ctx *Context, r game.OptionPicked) error {
		chosen := uuid.Nil
		if r.Option != nil {
			chosen = r.Option.Player
		}
		return body(ctx, chosen, r.Carry)
	})
}

func (c ChoosePlayerKeyed) Apply(ctx *Context) error {
	chooser := c.Chooser
	if chooser == uuid.Nil {
		chooser = ctx.Controller()
	}
	carry := append([]uuid.UUID{ctx.Controller()}, c.Carry...)
	queued := ctx.Game.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
		Chooser:  chooser,
		Among:    exceptPlayers(c.Among.seats(ctx, chooser), c.Except),
		Source:   ctx.Source(),
		Question: c.Question,
		ThenKey:  c.Then,
		Carry:    carry,
	})
	if queued != uuid.Nil {
		return nil
	}
	// Nothing could be asked: the continuation still runs, told nobody was
	// chosen (#544).
	return ctx.Game.RunOptionPickThenForEffect(c.Then, game.OptionPicked{
		Chooser: chooser, Source: ctx.Source(), Index: game.NoChoiceIndex, Carry: carry,
	})
}

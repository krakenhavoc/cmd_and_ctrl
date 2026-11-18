package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hour of Eternity — Sorcery {X}{X}{U}{U}{U}:
//
//	"Exile X target creature cards from your graveyard. For each card
//	 exiled this way, create a token that's a copy of that card, except
//	 it's a 4/4 black Zombie."
//
// Exactly X targets (CR 107.3a), exiled together; the copies are made
// from the cards that really reached exile, all at once, each a 4/4 black
// Zombie in place of its printed colours and creature types (The Scarab
// God's exception, CR 707.9b). The copy template is read from the exiled
// card, so a card that was exiled before its copy is made is still copied.
//
// No simplification.
func init() {
	targets := TargetCardInGraveyard("X target creature cards from your graveyard", Creature(), YouOwn())
	targets.CountFromX = true
	Register(Spec{
		OracleID:     "6fbaac3b-37ec-4815-a4ea-544c547e90b1",
		Name:         "Hour of Eternity",
		Completeness: CompletenessFull,
		Targets:      targets,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileTargetCardsThen(ctx, hourOfEternityCopies)
		},
	})
}

var hourOfEternityCopies = GraveyardExileTail("graveyard-exile/hour-of-eternity-zombie-copies",
	func(ctx *Context, exiled []uuid.UUID, _ map[uuid.UUID]bool) error {
		if len(exiled) == 0 {
			return nil
		}
		return CreateTokenCopiesTogether{
			Copies: exiled,
			Except: scarabGodZombieExceptionFor(ctx.Game.TextReaderForEffect(ctx.Item)),
		}.Apply(ctx)
	})

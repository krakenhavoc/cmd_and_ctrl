package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shadow Prophecy — Instant {2}{B}:
//
//	"Domain — Look at the top X cards of your library, where X is the
//	 number of basic land types among lands you control. Put up to two of
//	 them into your hand and the rest into your graveyard. You lose 2
//	 life."
//
// X is the shared domain count (DomainCount, #5666), counted as the spell
// resolves. The look is private; the player takes up to two (none is
// legal) and the rest go to the graveyard (TakeRestIntoGraveyard, which
// routes so a replacement sees them), then they lose 2 life whether or not
// anything was looked at.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "620f3851-493d-496e-880a-4a9dee074e79",
		Name:         "Shadow Prophecy",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			n := DomainCount(ctx.Game, player)
			if n <= 0 {
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), player, -2)
			}
			return TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, n),
				Max:      2,
				Optional: true,
				Label:    "Shadow Prophecy — put up to two of them into your hand and the rest into your graveyard",
				Key:      shadowProphecyTake,
			}.Apply(ctx)
		},
	})
}

// shadowProphecyTake is the take: the rest into the graveyard, then "you
// lose 2 life". An on-disk key: never renamed, never reused.
var shadowProphecyTake = RegisterTake("choose-cards/s82-dmu-shadow-prophecy-take",
	TakeDef{Then: func(g *game.Game, res TakeFromLibraryResult) error {
		if err := TakeRestIntoGraveyard(g, res); err != nil {
			return err
		}
		return g.ChangePlayerLifeForEffect(res.Source, res.Player, -2)
	}})

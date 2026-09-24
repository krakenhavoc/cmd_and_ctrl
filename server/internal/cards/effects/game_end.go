package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// game_end.go — the catalog's side of ADR 0057 (#749): winning and
// losing the game by an effect, and the "can't lose" / "can't win"
// gates a card declares.
//
// THE ONE RULE for a card that wins or loses the game: call WinTheGame
// or LoseTheGame and RETURN ITS ERROR. Never call the rotation or the
// game-over check yourself, and never swallow the error — it is
// game.ErrStopResolution when the resolution must not go on (the game
// ended, or the resolving item's own controller left the game, CR
// 800.4a), and returning it is what stops the rest of the effect.
// The engine treats that error as a clean stop, not a failure.

// WinTheGame is CR 104.2b: "<player> wins the game." The win is
// immediate, during the resolution (ADR 0057 Decision 3). A player who
// can't win — an opponent controls a Platinum Angel — doesn't, the
// engine logs the prevented win, and the resolution continues.
//
// Player defaults to the resolving item's controller ("you win the
// game").
type WinTheGame struct {
	Player uuid.UUID
}

// Apply implements Primitive.
func (w WinTheGame) Apply(ctx *Context) error {
	player := w.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	if _, err := ctx.Game.WinTheGameForEffect(player, ctx.Source()); err != nil {
		return err
	}
	if ctx.Game.State != game.StateActive {
		return game.ErrStopResolution
	}
	return nil
}

// LoseTheGame is CR 104.3e: "<player> loses the game." The loss is
// immediate (ADR 0057 Decision 3): the player leaves before this
// returns, unless a "can't lose the game" gate stops it.
//
// It returns game.ErrStopResolution when the game ended, or when the
// player who lost controls the resolving spell or ability — CR 800.4a
// makes the resolving object one of theirs, so the rest of it doesn't
// happen. When anyone else loses (Strixhaven Stadium's "that player
// loses the game"), the resolution continues without them.
//
// Player defaults to the resolving item's controller ("you lose the
// game").
type LoseTheGame struct {
	Player uuid.UUID
}

// Apply implements Primitive.
func (l LoseTheGame) Apply(ctx *Context) error {
	player := l.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	lost, err := ctx.Game.LoseTheGameForEffect(player, ctx.Source())
	if err != nil {
		return err
	}
	if ctx.Game.State != game.StateActive || (lost && player == ctx.Controller()) {
		return game.ErrStopResolution
	}
	return nil
}

// YouCantLoseOpponentsCantWin is Platinum Angel's and Herald of
// Eternal Dawn's printed static — "You can't lose the game and your
// opponents can't win the game." — as the two gates it is.
func YouCantLoseOpponentsCantWin() []game.GameEndGate {
	return []game.GameEndGate{
		{Scope: game.GateYou, CantLose: true},
		{Scope: game.GateOpponents, CantWin: true},
	}
}

// YouCantWinOpponentsCantLose is Abyssal Persecutor's — "You can't win
// the game and your opponents can't lose the game." — the mirror
// image.
func YouCantWinOpponentsCantLose() []game.GameEndGate {
	return []game.GameEndGate{
		{Scope: game.GateYou, CantWin: true},
		{Scope: game.GateOpponents, CantLose: true},
	}
}

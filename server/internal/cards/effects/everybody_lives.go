package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Everybody Lives! — Instant {1}{W} (Doctor Who):
//
//	"All creatures gain hexproof and indestructible until end of turn.
//	 Players gain hexproof until end of turn. Players can't lose life this
//	 turn and players can't lose the game or win the game this turn."
//
// Every creature on the battlefield as the spell resolves gets the two
// keywords until end of turn (a keyword grant locks its set, CR 611.2c);
// each player gains hexproof; "can't lose life" is game.ModCantLoseLife (CR
// 119.8: damage is still dealt, only the life loss does not happen, and a
// life payment can't be made); and the game-end gate stops every player from
// losing or winning (CR 104.2, 104.3) for the turn.
//
// No simplification.
func init() {
	const name = "Everybody Lives!"
	Register(Spec{
		OracleID:     "39213de3-6a4a-4879-a7f9-70f45013765e",
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerSurvive | game.AnswerProtect},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			g := ctx.Game
			if err := (GrantKeywordUntilEOT{Match: Creature(), Keywords: []string{"hexproof", "indestructible"},
				Label: name + " — hexproof and indestructible until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			d := DurationUntilEndOfTurn(ctx)
			for _, p := range g.Seats {
				if p == nil || p.Eliminated {
					continue
				}
				if err := (GainPlayerKeyword{Player: p.ID, Keyword: KeywordHexproof,
					Label: name + " — players gain hexproof until end of turn", Duration: d}).Apply(ctx); err != nil {
					return err
				}
			}
			g.PlayersCantLoseLifeForEffect(item.SourceCardID, d, name+" — players can't lose life this turn")
			g.GrantGameEndGateForEffect(item.Controller,
				game.GameEndGate{Scope: game.GateEachPlayer, CantLose: true, CantWin: true},
				name+" — players can't lose the game or win the game this turn", item.SourceCardID, d)
			return nil
		},
	})
}

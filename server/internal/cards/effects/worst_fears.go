package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Worst Fears — Sorcery {7}{B}:
//
//	"You control target player during that player's next turn. Exile
//	 Worst Fears. (You see all cards that player could see and make all
//	 decisions for them.)"
//
// Turn control (CR 723.1, ADR 0167, #3857), the plain form with no
// extra turn afterwards. The exile is the spell's last instruction,
// exactly as Temporal Mastery's is, so a countered Worst Fears goes to
// the graveyard like any other spell, and one whose only target left
// fizzles (CR 608.2b) and goes there too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "face0a43-6604-4508-b511-81ab0bef7b18",
		Name:         "Worst Fears",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := controlTargetPlayerNextTurn("Worst Fears")(ctx.Game, item); err != nil {
				return err
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}

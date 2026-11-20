package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Empyrial Storm — Sorcery {4}{W}{W} (Commander 2018, #5526):
//
//	"When you cast this spell, copy it for each time you've cast your
//	 commander from the command zone this game.
//	 Create a 4/4 white Angel creature token with flying."
//
// Echo Storm's cast trigger (echoStormCopies): one copy per cast of a
// commander from the command zone, summed over every commander you have and
// counted as the trigger resolves, made from last-known information so a
// countered original still leaves its copies. Each copy and the original
// make one Angel.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0af5f9f4-04d6-4701-be1f-65d4a0940737",
		Name:         "Empyrial Storm",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("Empyrial Storm — copy it for each time you've cast your commander from the command zone", echoStormCopies),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: item.Controller, Template: TokenCard("4/4 white Angel with flying"), N: 1}.Apply(ctx)
		},
	})
}

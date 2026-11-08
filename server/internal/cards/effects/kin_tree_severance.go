package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kin-Tree Severance — Instant {2/W}{2/B}{2/G} (Tarkir: Dragonstorm,
// #4463):
//
//	"Exile target permanent with mana value 3 or greater."
//
// Despark's floor at 3. ManaValueGE counts {X} as 0 on the battlefield
// (CR 202.3e), and a land (mana value 0) is never a legal target. Each
// {2/C} is paid with that colour or two mana of any kind (CR 107.4e), and
// the mana value is 6 (CR 202.3f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "211d1c59-0111-48dd-9ff6-d03e7ff2ad5c",
		Name:         "Kin-Tree Severance",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target permanent with mana value 3 or greater", ManaValueGE(3)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}

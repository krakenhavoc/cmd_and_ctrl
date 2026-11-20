package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hornswoggle — Instant {2}{U} (Rivals of Ixalan, #5531):
//
//	"Counter target creature spell. You create a Treasure token. (It's an
//	 artifact with "{T}, Sacrifice this token: Add one mana of any
//	 color.")"
//
// Essence Scatter followed by Spell Swindle's Treasure: a creature spell
// that can't be countered still pays out the Treasure, and a target that
// is gone by resolution makes the whole spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71b178bd-3b37-4de9-97a3-96e23d5444a4",
		Name:         "Hornswoggle",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpell("target creature spell", Creature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := counterTheTargetSpell(item, ctx); err != nil {
				return err
			}
			return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
		},
	})
}

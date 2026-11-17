package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Assault // Battery — split card (Invasion, #5616):
//
//	Assault — Sorcery {R}: "Assault deals 2 damage to any target."
//	Battery — Sorcery {3}{G}: "Create a 3/3 green Elephant creature
//	          token."
//
// Each half is an ADR 0034 face (ADR 0103): Assault under the bare
// oracle ID, Battery under "#1". Only one half can be cast (CR 709.3;
// the card has no fuse). Assault is Lightning Bolt's damageToFirstTarget
// at 2; Battery makes the shared "3/3 green Elephant" token row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fcdeb83f-531c-4087-9254-7f19eb065f00",
		Name:         "Assault",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		Targets:      TargetAny(),
		OnResolve:    damageToFirstTarget(2),
	})
	Register(Spec{
		OracleID:     "fcdeb83f-531c-4087-9254-7f19eb065f00#1",
		Name:         "Battery",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("3/3 green Elephant"), N: 1}.Apply(ctx)
		},
	})
}

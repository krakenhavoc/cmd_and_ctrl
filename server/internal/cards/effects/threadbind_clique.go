package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Threadbind Clique // Rip the Seams — Creature — Faerie {3}{U}, 3/3 //
// Instant — Adventure {2}{W} (Wilds of Eldraine, #4471):
//
//	Threadbind Clique — "Flying"
//	Rip the Seams — "Destroy target tapped creature. (Then exile this
//	 card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). Rip the Seams is Eriette's Lullaby's "target tapped
// creature" without the life: the target is checked again as the spell
// resolves (CR 608.2b), so a creature that untapped in response is not
// destroyed. The Adventure lifecycle is the engine's (CR 715.3d).
//
// No simplification.
const threadbindCliqueOracleID = "bd575e82-99e7-44c7-ab93-d33f5678e1ad"

func init() {
	Register(Spec{
		OracleID:        threadbindCliqueOracleID,
		Name:            "Threadbind Clique",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
	Register(Spec{
		OracleID:     threadbindCliqueOracleID + "#1",
		Name:         "Rip the Seams",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target tapped creature", tappedPermanent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return destroyFirstLegalTarget(ctx.Game, item)
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vraska's Fall — Instant {2}{B}: "Each opponent sacrifices a creature or
// planeswalker of their choice and gets a poison counter."
//
// The edict asks each opponent; the poison counters are its keyed
// continuation, run once every opponent has sacrificed. Every opponent gets
// the counter, whether or not they had anything to sacrifice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "33fd9192-3e9b-4d3a-92aa-06de29d449df",
		Name:         "Vraska's Fall",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return EachPlayerSacrifices{
				ExceptController: true,
				Match:            Or(Creature(), Planeswalker()),
				Label:            "a creature or planeswalker",
				Keyed:            s82OneB04Edicted,
			}.Apply(ctx)
		},
	})
}

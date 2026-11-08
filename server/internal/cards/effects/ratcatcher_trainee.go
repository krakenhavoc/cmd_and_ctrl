package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ratcatcher Trainee // Pest Problem — Creature — Human Peasant {1}{R},
// 2/1 // Instant — Adventure {2}{R} (Wilds of Eldraine, #4471):
//
//	Ratcatcher Trainee — "During your turn, this creature has first
//	 strike."
//	Pest Problem — "Create two 1/1 black Rat creature tokens with "This
//	 token can't block." (Then exile this card. You may cast the creature
//	 later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The first strike is Restless Spire's layer 6 grant on the
// controller's own turn, read at every layer pass, so it blinks off on an
// opponent's turn. Pest Problem makes two of the catalog's can't-block
// Rats. The Adventure lifecycle is the engine's (CR 715.3d).
//
// No simplification.
const ratcatcherTraineeOracleID = "ecc91e38-90fa-4d89-b262-d5f36dce5be4"

func init() {
	Register(Spec{
		OracleID:     ratcatcherTraineeOracleID,
		Name:         "Ratcatcher Trainee",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{rfReprintBFirstStrikeWhileCreatureOnYourTurn()},
	})
	Register(Spec{
		OracleID:     ratcatcherTraineeOracleID + "#1",
		Name:         "Pest Problem",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: ctx.Controller(), Template: CantBlockRatToken(), N: 2}.Apply(ctx)
		},
	})
}

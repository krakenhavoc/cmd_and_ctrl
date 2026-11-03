package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Return to the Sewers — {3}{U} Instant (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"Target creature's owner puts it on their choice of the top or bottom
//	 of their library. You create a Mutagen token. (It's an artifact with
//	 "{1}, {T}, Sacrifice this token: Put a +1/+1 counter on target
//	 creature. Activate only as a sorcery.")"
//
// Uneasy Partings' tuck: the owner, not the caster, picks top or bottom
// (PutInLibraryInAnyOrder with LibraryPlaceTopOrBottom). The token waits
// for that answer (CR 608.2c) and is the caster's. A target gone before
// resolution means the spell does not resolve (CR 608.2b), and no token.
// The Mutagen token is the catalog's CR 111.10v template (MutagenToken, #4676).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a3ee8565-f34c-4073-bc94-3fe9ec6bbe8a",
		Name:         "Return to the Sewers",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(returnToTheSewersTuck),
			CreateToken{Template: MutagenToken(), N: 1},
		),
	})
}

// returnToTheSewersTuck asks the target's owner top or bottom.
func returnToTheSewersTuck(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok || c.Owner == uuid.Nil {
			continue
		}
		return PutInLibraryInAnyOrder{
			Chooser:   c.Owner,
			Cards:     []uuid.UUID{t.ID},
			From:      game.ZoneBattlefield,
			Placement: game.LibraryPlaceTopOrBottom,
			Label:     "Return to the Sewers — put it on the top or bottom of your library",
		}.Apply(ctx)
	}
	return nil
}

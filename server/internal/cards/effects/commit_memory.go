package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commit // Memory — split card with aftermath, oracle
// 56fd8895-3be2-4591-86fa-87567d9cdc14:
//
//	Commit {3}{U} Instant: "Put target spell or nonland permanent into its
//	 owner's library second from the top."
//	Memory {4}{U}{U} Sorcery — Aftermath: "Each player shuffles their hand
//	 and graveyard into their library, then draws seven cards."
//
// Two specs, Memory under "<oracle_id>#1" as Dusk // Dawn is; aftermath is
// the engine's (CR 702.127a).
//
// Commit is Sudden Setback's target and move (a spell leaves the stack for
// the library without being countered, so a spell that can't be countered
// still goes) with a fixed depth instead of a choice: the placement is the
// resolving player's, since there is nothing to choose. Memory is Echo of
// Eons's body (echoOfEonsResolve).
//
// No simplification.
func init() {
	const oracle = "56fd8895-3be2-4591-86fa-87567d9cdc14"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Commit",
		Completeness: CompletenessFull,
		Targets:      TargetSpellOrPermanent("target spell or nonland permanent", nil, Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var id uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					id = t.ID
					break
				}
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || c.Owner == uuid.Nil {
				return nil
			}
			from := game.ZoneBattlefield
			if ctx.Game.StackItemForEffect(id) != nil {
				from = game.ZoneStack
			}
			return PutInLibraryInAnyOrder{
				Chooser:   ctx.Controller(),
				Cards:     []uuid.UUID{id},
				From:      from,
				Placement: game.LibraryPlaceTop,
				TopDepth:  2,
				Label:     "Commit — put it into its owner's library second from the top",
			}.Apply(ctx)
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Memory",
		Completeness: CompletenessFull,
		OnResolve:    echoOfEonsResolve,
	})
}

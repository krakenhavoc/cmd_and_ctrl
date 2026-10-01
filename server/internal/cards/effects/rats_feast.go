package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rats' Feast — Sorcery {X}{B}:
//
//	"Exile X target cards from a single graveyard."
//
// #1807, ADR 0106 §5. "X target cards" is an exact count bound to the
// announced X (CountFromX without UpToX, CR 601.2c: the number of
// targets is announced before the targets), and every one of them
// must come from one graveyard. So X can be no larger than the
// biggest graveyard at the table, which is the enumerator's reading
// too: it builds each X's sets inside one graveyard.
//
// No XMatters: the body reads the targets, not X, and the enumerator
// already never offers X=0 for a clause counted by X (its floor is 1).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "327c5ca2-4eeb-49cb-84fb-f6a4b06b24bc",
		Name:         "Rats' Feast",
		Completeness: CompletenessFull,
		Targets:      ratsFeastTargets(),
		OnResolve:    exileTargetCardsOnResolve,
	})
}

// ratsFeastTargets is "X target cards from a single graveyard".
func ratsFeastTargets() *game.TargetSpec {
	spec := TargetCardInGraveyard("X target cards from a single graveyard")
	spec.CountFromX = true
	return spec.AllShare(FromASingleGraveyard())
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Relentless Assault — Sorcery {2}{R}{R}:
//
//	"Untap all creatures that attacked this turn. After this main
//	 phase, there is an additional combat phase followed by an
//	 additional main phase."
//
// CR 500.8 through the turn plan (ADR 0059 sub-PR 2b, #753). The
// phases go on the CURRENT turn, and only when the spell resolves in a
// main phase (the 2025-09-19 ruling): cast in an opponent's main phase
// with a flash enabler, it gives that opponent the combat. "Creatures
// that attacked this turn" is the per-object attack history, so a
// creature that attacked and was then blinked is not untapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc1c0e8c-d0c1-445c-968f-7dec91e5d5fc",
		Name:         "Relentless Assault",
		Completeness: CompletenessFull,
		OnResolve:    untapAttackersThenCombatAndMain,
	})
}

// untapAttackersThenCombatAndMain is the whole of Relentless Assault,
// and the same sentence other cards print (Fury of the Horde, Waves of
// Aggression).
func untapAttackersThenCombatAndMain(_ *game.StackItem, ctx *Context) error {
	if err := (UntapCreaturesThatAttackedThisTurn{}).Apply(ctx); err != nil {
		return err
	}
	return ExtraCombatAndMainAfterThisMain().Apply(ctx)
}

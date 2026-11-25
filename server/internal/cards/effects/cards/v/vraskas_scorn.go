package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vraska's Scorn — Sorcery {2}{B}{B}:
//
//	"Target opponent loses 4 life. You may search your library and/or
//	 graveyard for a card named Vraska, Scheming Gorgon, reveal it, and put it
//	 into your hand. If you search your library this way, shuffle."
//
// The life loss is a life change from this spell to the target opponent (not
// damage). The search is Liliana's Scorn's: the named-card fetch of
// S82OldB105MayFetchNamed, a may-choice that then searches the library and/or
// graveyard and shuffles only when the library was searched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "020c26ba-ed46-47c3-925c-b9f32500f503",
		Name:         "Vraska's Scorn",
		Completeness: effects.CompletenessFull,
		Purpose:      s82OldB216ScornPurpose(),
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.TargetPlayerLosesLife{N: 4}.Apply),
			effects.Instruction(effects.S82OldB105MayFetchNamed("Vraska's Scorn", "Vraska, Scheming Gorgon")),
		),
	})
}

// s82OldB216ScornPurpose is the opponent's life loss plus the named fetch.
func s82OldB216ScornPurpose() game.Purpose {
	p := effects.ForTargets(game.TargetPurpose{Slot: 0, LifeLoss: 4})
	p.Tutors = 1
	return p
}

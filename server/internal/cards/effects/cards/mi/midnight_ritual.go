package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Midnight Ritual — Sorcery {X}{2}{B} (Mercadian Masques, S82 OLD batch 699, #5619):
//
//	"Exile X target creature cards from your graveyard. For each creature card
//	 exiled this way, create a 2/2 black Zombie creature token."
//
// Waste Management's shape: exactly X targets (CountFromX), each re-checked as
// the spell resolves (CR 608.2b), and a Zombie for each creature card that
// really reached exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "389ad507-8a14-4856-8d19-c90e3104aca8",
		Name:         "Midnight Ritual",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetsCountedByX(effects.TargetCardInGraveyard("X target creature cards from your graveyard",
			effects.YouOwn(), effects.Creature())),
		OnResolve: effects.S82OldB699MidnightRitual,
	})
}

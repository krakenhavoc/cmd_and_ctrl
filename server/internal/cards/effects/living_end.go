package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Living End — Sorcery, no mana cost (Secrets of Strixhaven Commander,
// #4834):
//
//	"Suspend 3—{2}{B}{B}
//	 Each player exiles all creature cards from their graveyard, then
//	 sacrifices all creatures they control, then puts all cards they
//	 exiled this way onto the battlefield."
//
// Living Death with no mana cost, so the way to play it is suspend (CR
// 702.62a): pay {2}{B}{B} at sorcery speed to exile it with three time
// counters, and it is cast for free when the last one comes off at the
// beginning of your upkeep. A card with no mana cost can't be cast
// otherwise (CR 202.3a, 118.6), and the engine's suspend (suspend.go) is
// what offers and counts it, as for Ancestral Vision. The resolution is
// Living Death's, step for step (livingDeathResolve): one exile of every
// graveyard's creature cards, one sacrifice of every creature on the
// battlefield, one simultaneous return of the cards exiled in the first
// step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3939fbe0-3be4-41ef-828d-c93f2cb33b4f",
		Name:         "Living End",
		Completeness: CompletenessFull,
		SpecialActions: []game.SpecialAction{
			Suspend(3, "{2}{B}{B}"),
		},
		OnResolve: livingDeathResolve,
	})
}

package effects

// Final Act — Sorcery {4}{B}{B} (EDHREC rank 3746):
//
//	"Choose one or more —
//	 • Destroy all creatures.
//	 • Destroy all planeswalkers.
//	 • Destroy all battles.
//	 • Exile all graveyards.
//	 • Each opponent loses all counters."
//
// Farewell in black. "Choose one or more" is ChooseN with Min 1 and
// Max 5 — every subset of the five is a legal cast — and the chosen
// modes resolve in printed order (CR 608.2c). The three sweeps go
// through the mass destroy, which drops indestructible permanents
// before the move (CR 702.12b) and fires the survivors' dies
// triggers as one batch; the fourth walks every seat's graveyard,
// the caster's included, through the helper Bojuka Bog and Farewell
// use; the fifth removes every counter of every kind — poison,
// energy, experience, rad, anything — from each opponent, the
// caster's own counters untouched.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "335a6c6a-030f-45d4-806c-467a22962eed",
		Name:         "Final Act",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or more", 1, 5,
			Mode("Destroy all creatures."),
			Mode("Destroy all planeswalkers."),
			Mode("Destroy all battles."),
			Mode("Exile all graveyards."),
			Mode("Each opponent loses all counters."),
		),
		OnResolve: b35FinalAct,
	})
}

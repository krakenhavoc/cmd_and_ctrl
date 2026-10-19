package effects

// Tezzeret's Gambit — Sorcery {3}{U/P} (EDHREC top 1,000, #3654):
//
//	"({U/P} can be paid with either {U} or 2 life.)
//	 Draw two cards, then proliferate."
//
// Two draws and then a proliferate, in printed order (CR 608.2c). The
// Phyrexian mana symbol is part of the printed cost and is paid by the
// engine's hybrid-Phyrexian cost handling like Zopandrel's {G/P}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b33d2b77-f068-4fdc-a1a8-0353a35e3973",
		Name:         "Tezzeret's Gambit",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			// The proliferate waits for a prompt the draw asks (CR 608.2c, #4155).
			DrawCards{N: 2},
			Proliferate{},
		),
	})
}

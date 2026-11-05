package effects

// Q Continuum — Land (Star Trek, #4973):
//
//	"This land enters tapped unless you control two or more other lands.
//	 When this land enters untapped, choose one —
//	 • Target creature gets +1/+1 and gains vigilance until end of turn.
//	 • Surveil 1.
//	 {T}: Add {G}."
//
// The green member of Star Trek's surveil-land cycle (trekB29SurveilLand).
//
// No simplification.
func init() {
	trekB29SurveilLand("457cf958-8129-4bf6-b4d5-211eb8c58763", "Q Continuum", "G",
		trekB29TargetCreatureUntilEOT("Target creature gets +1/+1 and gains vigilance until end of turn.", 1, 1, "vigilance"))
}

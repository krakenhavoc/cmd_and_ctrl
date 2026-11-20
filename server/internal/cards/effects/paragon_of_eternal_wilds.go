package effects

// Paragon of Eternal Wilds — Creature — Human Druid {3}{G}, 2/2 (Magic
// 2015, #5556):
//
//	"Other green creatures you control get +1/+1.
//	 {G}, {T}: Another target green creature you control gains trample
//	 until end of turn."
//
// The M15 Paragon cycle's shared body (m15B1Paragon).
//
// No simplification.
func init() {
	Register(m15B1Paragon("942a461a-49aa-4eeb-b399-a2afba41ade9", "Paragon of Eternal Wilds", "{G}", "G", "green", "trample"))
}

package effects

// Paragon of Open Graves — Creature — Skeleton Warrior {3}{B}, 2/2 (Magic
// 2015, #5556):
//
//	"Other black creatures you control get +1/+1.
//	 {2}{B}, {T}: Another target black creature you control gains
//	 deathtouch until end of turn."
//
// The M15 Paragon cycle's shared body (m15B1Paragon).
//
// No simplification.
func init() {
	Register(m15B1Paragon("39375b5d-f509-4482-8d34-797dae17d38b", "Paragon of Open Graves", "{2}{B}", "B", "black", "deathtouch"))
}

package effects

// Mirkwood — Land (The Hobbit, #4454):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {G}.
//	 {2}{B}{G}, {T}, Sacrifice this land: Put two +1/+1 counters on target
//	 Bear, Spider, or Wolf you control. Activate only as a sorcery."
//
// The body is shared with Elvenking's Halls and Lake-town
// (hobA6CounterLand). The target is any permanent you control with one of
// the three types, read off its effective subtypes.
//
// No simplification.
func init() {
	Register(hobA6CounterLand("cd49aa99-bf84-4edd-aecc-6dae78b73412", "Mirkwood", "BG",
		"Bear, Spider, or Wolf", "Bear", "Spider", "Wolf"))
}

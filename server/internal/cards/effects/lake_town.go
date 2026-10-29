package effects

// Lake-town — Land (The Hobbit, #4454):
//
//	"This land enters tapped.
//	 {T}: Add {W} or {U}.
//	 {2}{W}{U}, {T}, Sacrifice this land: Put two +1/+1 counters on target
//	 Human you control. Activate only as a sorcery."
//
// The body is shared with Mirkwood and Elvenking's Halls
// (hobA6CounterLand).
//
// No simplification.
func init() {
	Register(hobA6CounterLand("717c6beb-81c6-43ed-aab0-aedfc1cbac33", "Lake-town", "WU",
		"Human", "Human"))
}

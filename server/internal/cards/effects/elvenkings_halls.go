package effects

// Elvenking's Halls — Land (The Hobbit, #4454):
//
//	"This land enters tapped.
//	 {T}: Add {G} or {U}.
//	 {2}{G}{U}, {T}, Sacrifice this land: Put two +1/+1 counters on target
//	 Elf you control. Activate only as a sorcery."
//
// The body is shared with Mirkwood and Lake-town (hobA6CounterLand).
//
// No simplification.
func init() {
	Register(hobA6CounterLand("a91e0154-14a9-4681-8236-04db231592a4", "Elvenking's Halls", "GU",
		"Elf", "Elf"))
}

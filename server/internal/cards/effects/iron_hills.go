package effects

// Iron Hills — Land (The Hobbit):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {W}.
//	 {2}{R}{W}, {T}, Sacrifice this land: Put two +1/+1 counters on target
//	 Dwarf you control. Activate only as a sorcery."
//
// The shared two-colour counter land (hobB7CounterLand). No simplification.
func init() {
	Register(hobB7CounterLand("a71e8d07-1a49-47a0-834e-de87d750a200", "Iron Hills",
		"{R|W}", "Add {R} or {W}", "{2}{R}{W}", "target Dwarf you control", Subtype("Dwarf")))
}

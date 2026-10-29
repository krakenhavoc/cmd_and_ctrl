package effects

// Goblin-town — Land (The Hobbit):
//
//	"This land enters tapped.
//	 {T}: Add {B} or {R}.
//	 {2}{B}{R}, {T}, Sacrifice this land: Put two +1/+1 counters on target
//	 Goblin or Orc you control. Activate only as a sorcery."
//
// The shared two-colour counter land (hobB7CounterLand). No simplification.
func init() {
	Register(hobB7CounterLand("98d20908-6d68-4d31-b719-207f93c9b402", "Goblin-town",
		"{B|R}", "Add {B} or {R}", "{2}{B}{R}", "target Goblin or Orc you control", Or(Subtype("Goblin"), Subtype("Orc"))))
}

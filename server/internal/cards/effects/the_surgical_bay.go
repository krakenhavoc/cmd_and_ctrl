package effects

// The Surgical Bay — Land — Sphere (Phyrexia: All Will Be One, #5479):
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 {1}{U}, {T}, Sacrifice this land: Draw a card."
//
// The tapped entry is a real CR 614 self-replacement, the mana a mana
// ability, and the draw an activated ability whose cost taps and
// sacrifices the land (CR 602.2).
//
// No simplification.
func init() {
	Register(s82OneSphereDrawLand("db9c155f-b342-41bb-9e1a-50358fb9f40e", "The Surgical Bay", "U"))
}

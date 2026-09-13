package effects

// Sunken Ruins — Land (EDHREC rank 983):
//
//	"{T}: Add {C}.
//	 {U/B}, {T}: Add {U}{U}, {U}{B}, or {B}{B}."
//
// The Dimir member of the Shadowmoor filter-land cycle Graven Cairns
// established (batch 04), through the b08FilterLand helper: the
// hybrid cost paid from the pool as printed, and two independent
// {U|B} picks for the three-way output. See graven_cairns.go for the
// reasoning; nothing here differs but the colours.
//
// No simplification.
func init() {
	Register(b08FilterLand("e6415ffb-8b7a-41c3-bedf-0d4112b7b795", "Sunken Ruins", "U", "B"))
}

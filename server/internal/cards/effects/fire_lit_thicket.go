package effects

// Fire-Lit Thicket — Land (EDHREC rank 1206):
//
//	"{T}: Add {C}.
//	 {R/G}, {T}: Add {R}{R}, {R}{G}, or {G}{G}."
//
// The Gruul member of the Shadowmoor filter-land cycle Graven Cairns
// established (batch 04), through the b08FilterLand helper: the
// hybrid cost paid from the pool as printed, and two independent
// {R|G} picks for the three-way output. See graven_cairns.go for the
// reasoning; nothing here differs but the colours.
//
// No simplification.
func init() {
	Register(b08FilterLand("d99a1d9a-7721-4331-bf22-1c6ee0bd825a", "Fire-Lit Thicket", "R", "G"))
}

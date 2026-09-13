package effects

// Wooded Bastion — Land (EDHREC rank 2028):
//
//	"{T}: Add {C}.
//	 {G/W}, {T}: Add {G}{G}, {G}{W}, or {W}{W}."
//
// The Selesnya member of the Shadowmoor filter-land cycle Graven
// Cairns established (batch 04), through the b08FilterLand helper:
// the hybrid cost paid from the pool as printed, and two independent
// {G|W} picks for the three-way output. See graven_cairns.go for the
// reasoning; nothing here differs but the colours.
//
// No simplification.
func init() {
	Register(b08FilterLand("61b85077-64aa-4bcc-890d-2d88da9543c0", "Wooded Bastion", "G", "W"))
}

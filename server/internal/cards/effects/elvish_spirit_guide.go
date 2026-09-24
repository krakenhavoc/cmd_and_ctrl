package effects

// Elvish Spirit Guide — Creature — Elf Spirit {2}{G}, 2/2:
//
//	"Exile this card from your hand: Add {G}."
//
// Simian Spirit Guide one colour over, and the second half of
// #1228's proof: the primitive is a declaration on the ability, not a
// special case for one card, so the catalog entry is the same one
// line with a different mana symbol in it.
//
// It is the older of the two by twenty years (Alliances, 1996) and
// the reason the pair is worth having both of: a test that the green
// one taps for {G} and the red one for {R} is a test that
// ExileFromHandForMana reads its argument rather than the engine
// knowing the cards.
func init() {
	Register(Spec{
		OracleID:      "6b0e23cf-7d68-4329-86db-7adc26abd86b",
		Name:          "Elvish Spirit Guide",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{ExileFromHandForMana("{G}")},
	})
}

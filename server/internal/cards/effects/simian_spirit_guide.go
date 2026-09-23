package effects

// Simian Spirit Guide — Creature — Ape Spirit {2}{R}, 2/2:
//
//	"Exile this card from your hand: Add {R}."
//
// #1228's proof that a MANA ability can function from a zone other
// than the battlefield (CR 113.6). The body is a vanilla 2/2 and
// nobody casts it, so a test that sees {R} in the pool and the card
// in exile is seeing the primitive and nothing else.
//
// The Spec is one `ExileFromHandForMana("{R}")` entry and the zone
// dimension does the rest — the same shape Dregscape Zombie's unearth
// has one ability kind over. See spirit_guide.go, and
// game/mana_ability_zone.go for why the declaration and the
// exile-this cost imply each other at boot.
func init() {
	Register(Spec{
		OracleID:      "44e0ffa3-8915-4c1f-8f1a-4aeea1365f07",
		Name:          "Simian Spirit Guide",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{ExileFromHandForMana("{R}")},
	})
}

package effects

// Skinbrand Goblin — Creature — Goblin Warrior {1}{R}, 2/1 (Gatecrash, #5565):
//
//	"Bloodrush — {R}, Discard this card: Target attacking creature gets +2/+1
//	 until end of turn."
//
// Bloodrush is Pyrewild Shaman's shape (gtcBloodrush): an activated
// ability that works from the hand, with discarding this card as part of
// the cost. A target that stopped attacking or left is illegal (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e1d5873-50e6-432a-b7a9-9f8d2632a0b9",
		Name:         "Skinbrand Goblin",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{gtcBloodrush("Skinbrand Goblin", "{R}", 2, 1, "")},
	})
}

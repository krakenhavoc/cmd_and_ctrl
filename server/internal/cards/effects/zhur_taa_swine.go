package effects

// Zhur-Taa Swine — Creature — Boar {3}{R}{G}, 5/4 (Gatecrash, #5565):
//
//	"Bloodrush — {1}{R}{G}, Discard this card: Target attacking creature gets
//	 +5/+4 until end of turn."
//
// Bloodrush is Pyrewild Shaman's shape (gtcBloodrush): an activated
// ability that works from the hand, with discarding this card as part of
// the cost. A target that stopped attacking or left is illegal (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dcc31cbf-f75d-4584-a97b-aa9d19d7b142",
		Name:         "Zhur-Taa Swine",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{gtcBloodrush("Zhur-Taa Swine", "{1}{R}{G}", 5, 4, "")},
	})
}

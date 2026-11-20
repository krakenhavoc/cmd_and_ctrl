package effects

// Scab-Clan Charger — Creature — Centaur Warrior {3}{G}, 2/4 (Gatecrash, #5565):
//
//	"Bloodrush — {1}{G}, Discard this card: Target attacking creature gets
//	 +2/+4 until end of turn."
//
// Bloodrush is Pyrewild Shaman's shape (gtcBloodrush): an activated
// ability that works from the hand, with discarding this card as part of
// the cost. A target that stopped attacking or left is illegal (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f947056-fa10-4491-95e5-532ea3b76d10",
		Name:         "Scab-Clan Charger",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{gtcBloodrush("Scab-Clan Charger", "{1}{G}", 2, 4, "")},
	})
}

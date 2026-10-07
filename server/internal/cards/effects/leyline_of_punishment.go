package effects

// Leyline of Punishment — Enchantment {2}{R}{R}:
//
//	"If this card is in your opening hand, you may begin the game with
//	 it on the battlefield.
//	 Players can't gain life.
//	 Damage can't be prevented."
//
// Both statics are ADR 0107 §5's battlefield gates, read live while the
// Leyline is on the battlefield and has its abilities (CR 613.1f):
// CR 119.7's "can't gain life" for every player, and CR 615.12's
// "damage can't be prevented" for every damage event.
//
// The opening-hand clause (CR 103.6a) is Spec.OpeningHand (ADR 0133):
// the seat holding it is asked as the mulligan window closes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "2608df54-dfbe-417d-aef5-49afbdfb03da",
		Name:                  "Leyline of Punishment",
		Completeness:          CompletenessFull,
		OpeningHand:           BeginTheGameOnTheBattlefield(),
		CantGainLife:          PlayersCantGainLife(),
		DamageCantBePrevented: DamageCantBePreventedStatic(),
	})
}

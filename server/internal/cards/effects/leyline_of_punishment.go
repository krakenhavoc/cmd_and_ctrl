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
// ONE SIMPLIFICATION, the one every Leyline in the catalog carries
// (leyline_of_sanctity.go): there is no pre-game window in which a card
// in an opening hand can be put onto the battlefield, so it is cast for
// its four mana. Strictly weaker than printed.
func init() {
	Register(Spec{
		OracleID:     "2608df54-dfbe-417d-aef5-49afbdfb03da",
		Name:         "Leyline of Punishment",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Starting the game with this on the battlefield from your opening hand isn't implemented — you cast it for {2}{R}{R} like an ordinary enchantment.",
		},
		CantGainLife:          PlayersCantGainLife(),
		DamageCantBePrevented: DamageCantBePreventedStatic(),
	})
}

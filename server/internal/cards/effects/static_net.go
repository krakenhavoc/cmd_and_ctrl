package effects

// Static Net — Enchantment {3}{W} (The Brothers' War):
//
//	"When this enchantment enters, exile target nonland permanent an
//	 opponent controls until this enchantment leaves the battlefield.
//	 When this enchantment enters, you gain 2 life and create a tapped
//	 Powerstone token."
//
// Two printed triggers, so two registered rows. The first is Banishing
// Light's (CR 610.3): the exile records its own return. The second is
// independent of it; the Powerstone enters tapped because the effect
// says so.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f285781-ee03-4a58-a85e-a33f5efffc9d",
		Name:         "Static Net",
		Completeness: CompletenessFull,
		Triggered: append(exileUntilThisLeavesTriggers("Static Net"),
			WhenThisEnters("Static Net — you gain 2 life and create a tapped Powerstone token",
				Do(
					GainLife{Amount: 2},
					CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: 1},
				))),
	})
}

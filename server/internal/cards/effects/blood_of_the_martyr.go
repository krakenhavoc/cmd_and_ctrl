package effects

// Blood of the Martyr — Instant {W}{W}{W}:
//
//	"Until end of turn, if damage would be dealt to any creature, you may
//	 have that damage dealt to you instead."
//
// ADR 0108 §9, amendment 2026-10-16 (#2056): a redirection for the rest
// of the turn whose protected set is every creature, whoever controls it,
// read as the damage would be dealt (CR 611.2c, so a creature that enters
// later is covered), and whose "you may" is asked of the spell's
// controller each time damage would be dealt to a creature (CR 614.5:
// once per damage event). "You" is read as the spell resolves; once that
// player has left the game the redirection does nothing (CR 614.9).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "0e69e6c2-fdc5-4c1a-bd96-c4283b82743a",
		Name:         "Blood of the Martyr",
		Completeness: CompletenessFull,
		OnResolve:    redirectSpell(RedirectDamage{Protect: ShieldCreatures, To: RedirectToYou, Optional: true}),
	})
}

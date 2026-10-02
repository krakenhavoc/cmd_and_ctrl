package effects

// Mercenaries — Creature — Human Mercenary {3}{W}:
//
//	"{3}: The next time this creature would deal damage to you this turn, prevent that damage. Any player may activate this ability."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from one source, here the Mercenaries themselves (CR 615.8).
// It is an any-player ability (CR 602.2, ADR 0106 #1793), so "you" is
// whoever activated it. The shield names the Mercenaries as the object
// they were when activated (CR 400.7): one that dies and deals damage
// "as it last existed" is still them; one that comes back is not.
//
// No simplifications.
func init() {
	row := nextDamageShieldRow(
		"{3}: The next time this creature would deal damage to you this turn, prevent that damage. Any player may activate this ability.",
		ManaCost("{3}"), nil, PreventNextDamageFromThis(ShieldYou))
	row.AnyPlayer = true
	Register(Spec{
		OracleID:     "9d96671c-7a98-48ec-b479-bc451e18264d",
		Name:         "Mercenaries",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{row},
	})
}

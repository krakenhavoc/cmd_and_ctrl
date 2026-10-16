package effects

// Nova Pentacle — Artifact {4}:
//
//	"{3}, {T}: The next time a source of your choice would deal damage
//	 to you this turn, that damage is dealt to target creature of an
//	 opponent's choice instead."
//
// General's Regalia's redirection (ADR 0108 §9) with the target chosen
// by an opponent (#2055, ADR 0019's 2026-10-16 amendment): as the
// ability is activated its controller chooses an opponent, and that
// opponent chooses the creature (the rulings). It doesn't have to be a
// creature that opponent controls, and with no creature to target the
// ability can't be activated. The source is chosen as the ability
// resolves (CR 609.7a). In a multiplayer game a different opponent may
// be chosen each time.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "db121503-8a34-498a-829c-72c33798369b",
		Name:         "Nova Pentacle",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			redirectRow("{3}, {T}: The next time a source of your choice would deal damage to you this turn, that damage is dealt to target creature of an opponent's choice instead.",
				Plus(ManaCost("{3}"), TapCost()),
				OfAnOpponentsChoice(TargetCreature("target creature of an opponent's choice")),
				RedirectDamage{Choose: true, Protect: ShieldYou, Next: true, To: RedirectToClause(0)}),
		},
	})
}

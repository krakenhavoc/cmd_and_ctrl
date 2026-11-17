package effects

// Errant, Street Artist — Legendary Creature — Human Rogue {U}, 0/3
// (Streets of New Capenna Commander, #5491):
//
//	"Flash
//	 Defender, haste
//	 {1}{U}, {T}: Copy target spell you control that wasn't cast. You may
//	 choose new targets for the copy."
//
// Slick Imitator's copy body with one more condition on the target: a spell
// is "not cast" when it is a copy created on the stack (CR 707.10), so a
// Bolt you cast is refused and its copy is a legal target. The tap needs
// no summoning sickness (haste). The copy of a permanent spell becomes a
// token as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "23e6df25-90db-42a4-8658-e592e67be15e",
		Name:            "Errant, Street Artist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "defender", "haste"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, {T}: Copy target spell you control that wasn't cast. You may choose new targets for the copy.",
			Cost:    Plus(ManaCost("{1}{U}"), TapCost()),
			Targets: TargetSpell("target spell you control that wasn't cast", YouControl(), s82SncSpellWasNotCast),
			Effect:  copyTargetedSpell,
		}},
	})
}

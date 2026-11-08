package effects

// Deadly Plot — Instant {3}{B} (Foundations, #4465):
//
//	"Choose one —
//	 • Destroy target creature or planeswalker.
//	 • Return target Zombie creature card from your graveyard to the
//	   battlefield tapped."
//
// Molten Collapse's first bullet, and Phoenix Down's tapped return with
// the clause narrowed to a Zombie creature card you own. Each bullet has
// its own target, chosen as the spell is cast (CR 700.2a), and does
// nothing if that target is illegal as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6178715c-870d-4710-b758-66e080804ee3",
		Name:         "Deadly Plot",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Destroy target creature or planeswalker.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
				DestroyTheModesTarget),
			ModeDoing("Return target Zombie creature card from your graveyard to the battlefield tapped.",
				TargetCardInGraveyard("target Zombie creature card from your graveyard", YouOwn(), Creature(), HasSubtype("Zombie")),
				ReturnTheModesTargetToBattlefieldTapped),
		),
	})
}

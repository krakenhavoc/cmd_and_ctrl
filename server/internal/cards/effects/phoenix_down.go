package effects

// Phoenix Down — Artifact {W} (deck request #3655):
//
//	"{1}{W}, {T}, Exile this artifact: Choose one —
//	 • Return target creature card with mana value 4 or less from your
//	   graveyard to the battlefield tapped.
//	 • Exile target Skeleton, Spirit, or Zombie."
//
// Exiling the artifact is a cost, so it is gone by the time either mode
// resolves. Each mode's target is judged again on resolution
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8cc7e418-cc5e-4aef-9757-874223fccd74",
		Name:         "Phoenix Down",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}{W}, {T}, Exile this artifact: Choose one — return target creature card with mana value 4 or less from your graveyard to the battlefield tapped; or exile target Skeleton, Spirit, or Zombie.",
			Cost:  Plus(ManaCost("{1}{W}"), TapCost(), ExileThis()),
			Modes: ChooseOne(
				ModeDoing("Return target creature card with mana value 4 or less from your graveyard to the battlefield tapped.",
					TargetCardInGraveyard("target creature card with mana value 4 or less in your graveyard",
						YouOwn(), Creature(), ManaValueLE(4)),
					ReturnTheModesTargetToBattlefieldTapped),
				ModeDoing("Exile target Skeleton, Spirit, or Zombie.",
					TargetPermanent("target Skeleton, Spirit, or Zombie", AnySubtype("Skeleton", "Spirit", "Zombie")),
					exileTheModesTarget),
			),
		}},
	})
}

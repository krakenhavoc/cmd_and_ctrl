package effects

// Resistance Is Futile — Sorcery {1}{B} (Star Trek, #4973):
//
//	"Choose one —
//	 • Destroy target nonartifact creature.
//	 • Return target artifact creature card with mana value 3 or less
//	   from your graveyard to the battlefield."
//
// Each bullet's target is chosen as the spell is cast and re-checked as it
// resolves (CR 608.2b). A card with {X} in its cost counts X as 0 in the
// graveyard (CR 202.3e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "49975689-85f8-4221-bb11-3d3f68a1a7ac",
		Name:         "Resistance Is Futile",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Destroy target nonartifact creature.",
				TargetCreature("target nonartifact creature", Not(Artifact())),
				DestroyTheModesTarget),
			ModeDoing("Return target artifact creature card with mana value 3 or less from your graveyard to the battlefield.",
				TargetCardInGraveyard("target artifact creature card with mana value 3 or less from your graveyard",
					YouOwn(), Artifact(), Creature(), ManaValueLE(3)),
				ReturnTheModesTargetToBattlefield),
		),
	})
}

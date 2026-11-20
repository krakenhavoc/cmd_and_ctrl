package effects

// Recollect — Sorcery {2}{G} (Ravnica: City of Guilds, #5600):
//
//	"Return target card from your graveyard to your hand."
//
// Regrowth's body: any card in its caster's graveyard returns to hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b38fafb-3999-46dc-928b-1677625c1943",
		Name:         "Recollect",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target card from your graveyard", YouOwn()),
		OnResolve:    returnTargetFromGraveyardToHand,
	})
}

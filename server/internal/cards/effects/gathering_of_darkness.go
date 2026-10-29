package effects

// Gathering of Darkness — Sorcery {3}{B} (The Hobbit, HOB):
//
//	"Return up to one target creature card from your graveyard to your
//	 hand.
//	 Amass Goblins 3."
//
// With no target chosen the spell still resolves and amasses; with a target
// that has left the graveyard it does not resolve (CR 608.2b), as Fates'
// Reversal. The amass waits for any prompt the return asks (CR 608.2c, #4155).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "85e1556d-c4da-4979-a5db-4ce3842c7797",
		Name:         "Gathering of Darkness",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("up to one target creature card from your graveyard", YouOwn(), Creature()).WithCount(0, 1),
		OnResolve: ResolveSteps(
			Instruction(returnLegalGraveyardTargetsToHand),
			Amass{Subtype: "Goblin", N: 3},
		),
	})
}

package effects

// Macabre Waltz — Sorcery {1}{B} (Jumpstart, #4892):
//
//	"Return up to two target creature cards from your graveyard to your
//	 hand, then discard a card."
//
// The discard is asked after the returns (CR 608.2c), so a card just
// returned is a legal discard. With no target chosen the spell still
// resolves and you discard; when every chosen target has left, the spell
// does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf75dbd8-b65a-48d9-b96b-4afb43f336d1",
		Name:         "Macabre Waltz",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("up to two target creature cards from your graveyard", YouOwn(), Creature()).WithCount(0, 2),
		OnResolve: ResolveSteps(
			Instruction(returnLegalGraveyardTargetsToHand),
			Instruction(controllerDiscardsACard),
		),
	})
}

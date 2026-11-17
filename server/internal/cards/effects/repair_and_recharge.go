package effects

// Repair and Recharge — Sorcery {3}{W}{W} (The Brothers' War, #5482):
//
//	"Return target artifact, enchantment, or planeswalker card from your
//	 graveyard to the battlefield. Create a tapped Powerstone token."
//
// The one-target reanimation (Agadeem's Awakening's return), then the
// Powerstone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f42517b-e62e-4b8c-8c56-cb32fb62c478",
		Name:         "Repair and Recharge",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target artifact, enchantment, or planeswalker card from your graveyard",
			YouOwn(), Or(Artifact(), Enchantment(), Planeswalker())),
		OnResolve: ResolveSteps(
			Instruction(returnLegalGraveyardTargetsToBattlefield),
			CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: 1},
		),
	})
}

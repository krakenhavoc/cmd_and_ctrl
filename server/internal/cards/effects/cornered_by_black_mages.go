package effects

// Cornered by Black Mages — Sorcery {1}{B}{B} (EDHREC rank 3076):
//
//	"Target opponent sacrifices a creature of their choice.
//	 Create a 0/1 black Wizard creature token with "Whenever you
//	 cast a noncreature spell, this token deals 1 damage to each
//	 opponent.""
//
// An edict that leaves a body behind. The target is the opponent;
// the sacrifice is THEIR choice — PlayerSacrificesForEffect queues
// them a prompt over their own creatures, which is what "of their
// choice" means and why hexproof is irrelevant — and the Wizard is
// created in the same resolution (b29TargetOpponentSacrificesACreatureThenWizard).
//
// Sandbox simplification, declared, weaker than printed: the
// Wizard's own "whenever you cast a noncreature spell" ping is NOT
// implemented. A token template carries no triggered abilities and
// a token has no oracle ID for the catalog to key one on; the
// Synthesizer posture — the source carrying the token's ability on
// its behalf — is not available to a sorcery, which is in the
// graveyard by the time anything is cast. So the Wizard is a 0/1
// body and nothing more; the edict, which is the card, is whole.
func init() {
	Register(Spec{
		OracleID:     "68a58e0b-1506-492e-8ca3-016e520c10ba",
		Name:         "Cornered by Black Mages",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Wizard token is a plain 0/1 — it doesn't deal 1 damage to each opponent when you cast a noncreature spell."},
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve:    b29TargetOpponentSacrificesACreatureThenWizard,
	})
}

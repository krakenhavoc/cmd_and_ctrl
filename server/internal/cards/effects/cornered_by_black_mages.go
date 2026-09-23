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
// The Wizard's own "whenever you cast a noncreature spell" ping ships
// since ADR 0083 (#1248): it is the `token:wizard` catalog template
// (tokens.go, shared with Mysidian Elder), and its trigger is found
// through `game.CatalogKey`'s token-key fallback. This card is
// exactly why the ability has to live on the TOKEN — the Synthesizer
// posture, where the source carries the token's ability on its
// behalf, is not available to a sorcery, which is in the graveyard by
// the time anything is cast.
func init() {
	Register(Spec{
		OracleID:     "68a58e0b-1506-492e-8ca3-016e520c10ba",
		Name:         "Cornered by Black Mages",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve:    b29TargetOpponentSacrificesACreatureThenWizard,
	})
}

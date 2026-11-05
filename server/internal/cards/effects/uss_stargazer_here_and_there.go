package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// U.S.S. Stargazer, Here and There — {3} Legendary Artifact — Spacecraft
// 4/3 (Star Trek Commander, #4974):
//
//	"Station (… It's an artifact creature at 6+.)
//	 6+ | Flying
//	 Picard Maneuver — Whenever U.S.S. Stargazer attacks, create a tapped
//	 and attacking 4/3 colorless Spacecraft Illusion artifact creature
//	 token with flying. Exile that token at end of combat."
//
// The attack trigger is printed in the 6+ striation (CR 721.2), so it
// exists only from six charge counters on, as H.M.S. Bounty's combat
// trigger does. The token is Dorothea's Retribution's shape: made
// tapped and attacking what its controller chooses (CR 508.4), the
// Stargazer's own defender offered first, never declared so it fires no
// attack trigger, and exiled by a delayed trigger at the beginning of
// the end of combat step (CR 511.2). Picard Maneuver is an ability word
// (CR 207.2c).
//
// No simplification.
func init() {
	t := WheneverThisAttacks("U.S.S. Stargazer — Picard Maneuver: create a 4/3 Spacecraft Illusion with flying tapped and attacking; exile it at end of combat",
		trekB43StargazerIllusion)
	t.ActiveWhen = AtChargeCounters(6)
	Register(Spec{
		OracleID:     "a5b72c5d-4c73-49cd-9fc5-855856807ff2",
		Name:         "U.S.S. Stargazer, Here and There",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered:    []game.TriggeredAbility{t},
		Activated:    []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(6, 4, 3),
			ThresholdKeywords(6, "flying"),
		),
	})
}

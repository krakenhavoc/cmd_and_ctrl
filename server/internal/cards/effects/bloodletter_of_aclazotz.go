package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodletter of Aclazotz — Creature — Vampire Demon {1}{B}{B}{B}, 2/4
// (EDHREC rank 854):
//
//	"Flying
//	 If an opponent would lose life during your turn, they lose twice
//	 that much life instead. (Damage causes loss of life.)"
//
// A CR 614 replacement on the life window (RepEventLife), applied to a
// LOSS by an opponent while Bloodletter's controller is the active
// player. Every way a player loses life runs that window: a drain, a
// life payment (CR 119.4) and, since ADR 0169 (#3388), the life that
// damage costs (CR 120.3a). The damage itself is unchanged, so lifelink
// gains the damage dealt and the commander tally counts it; "whenever a
// player loses life" triggers see the doubled loss. Two Bloodletters
// quadruple. The Angel's Grace floor caps the doubled loss, and a
// locked life total loses nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "469956a2-7cd4-4695-b8f4-c841526f160d",
		Name:            "Bloodletter of Aclazotz",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			OpponentsLoseTwiceThatMuchLifeDuringYourTurn("Bloodletter of Aclazotz: lose twice that much life"),
		},
	})
}

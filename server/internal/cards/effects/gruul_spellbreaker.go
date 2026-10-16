package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gruul Spellbreaker — Creature — Ogre Warrior {1}{R}{G}, 3/3:
//
//	"Riot (This creature enters with your choice of a +1/+1 counter or
//	 haste.)
//	 Trample
//	 During your turn, you and this creature have hexproof."
//
// Riot is the engine's keyword (ADR 0109 §10, #1556). The last line is
// one static ability with two receivers, and the engine spells each in
// its own place:
//
//   - the creature's hexproof (CR 702.11b) is a layer-6 grant to itself
//     whose condition reads the turn, re-answered on every recompute;
//   - the player's (CR 702.11c) is Spec.PlayerKeywordsIf (#2013, ADR
//     0072's 2026-10-16 amendment), asked on every targeting query, so
//     an opponent's spell aimed at you on your turn is refused and one
//     cast on their own turn is not.
//
// "Your turn" is the controller's, read now: a Spellbreaker that changes
// control shields its new controller on their turns.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "827b3ed0-09b3-4d89-8b81-dd89cc002782",
		Name:             "Gruul Spellbreaker",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{game.KeywordRiot, "trample"},
		Static:           []game.StaticAbility{ThisHasDuringYourTurn(KeywordHexproof)},
		PlayerKeywordsIf: []game.PlayerKeywordIf{YouHaveDuringYourTurn(KeywordHexproof)},
	})
}

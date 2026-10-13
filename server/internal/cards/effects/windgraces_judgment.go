package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Windgrace's Judgment — Instant {3}{B}{G} (EDHREC rank 3173):
//
//	"For any number of opponents, destroy target nonland permanent
//	 that player controls."
//
// One removal spell per opponent at instant speed — the multiplayer
// Putrefy. The target clause is "any number of target nonland
// permanents your opponents control", and since #1559 the "one per
// opponent" is the clause's own set rule, EachDifferentController:
// the announce gate refuses two permanents of one opponent's
// (CR 601.2c), the picker greys the second, and the bot is never
// offered the pair. Before #1559 the picker accepted the pair and the
// second one was silently skipped as the spell resolved.
//
// The clause is one "up to one" clause per opponent in turn order
// (perOpponentClauses), each bound to its opponent: a permanent that
// changes control in response is an illegal target for the clause it
// was chosen under (CR 608.2b), even if it moved to another opponent.
func init() {
	Register(Spec{
		OracleID:     "de1ca6ed-b275-4f62-ba05-f31b3659b352",
		Name:         "Windgrace's Judgment",
		Completeness: CompletenessFull,
		Targets:      perOpponentClauses(game.MaxPlayers-1, true, "nonland permanent", Nonland()),
		OnResolve:    destroyEachLegalTarget,
	})
}

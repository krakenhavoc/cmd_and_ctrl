package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mogg Toady — Creature — Goblin {1}{R}, 2/2:
//
//	"This creature can't attack unless you control more creatures than
//	 defending player.
//	 This creature can't block unless you control more creatures than
//	 attacking player."
//
// The attack half is #1879's restriction (ADR 0107 §2, CR 508.1c): the
// defending player is worked out per target (CR 508.5, 508.5a), so in
// Commander it may attack only the opponents who control fewer creatures
// than you, with their planeswalkers and the battles they protect. The
// block half is a CR 509.1b block rule; the attacking player is the
// active player (CR 506.2). Both count creatures as they are when the
// creature is declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77c7ba06-d247-421d-9e0c-54c3000c1b78",
		Name:         "Mogg Toady",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{CantAttackUnlessYouControlMoreThanDefendingPlayer(QueryType("creature"))},
		BlockRules: []game.BlockRule{CantBlockUnless(YouControlMoreThanTheAttackingPlayer(QueryType("creature")),
			"it can't block unless its controller controls more creatures than the attacking player")},
	})
}

package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Monstrous Hound — Creature — Dog {3}{R}, 4/4:
//
//	"This creature can't attack unless you control more lands than
//	 defending player.
//	 This creature can't block unless you control more lands than
//	 attacking player."
//
// The attack half is #1879's restriction (ADR 0107 §2, CR 508.1c): the
// defending player is worked out per target (CR 508.5, 508.5a), so in
// Commander it may attack only the opponents who control fewer lands
// than you, with their planeswalkers and the battles they protect. The
// block half is a CR 509.1b block rule; the attacking player is the
// active player (CR 506.2). Both count lands as they are when the
// creature is declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97589fbb-3bd9-49c9-b8d8-bdc895a7f1e6",
		Name:         "Monstrous Hound",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.CantAttackUnlessYouControlMoreThanDefendingPlayer(effects.QueryType("land"))},
		BlockRules: []game.BlockRule{effects.CantBlockUnless(effects.YouControlMoreThanTheAttackingPlayer(effects.QueryType("land")),
			"it can't block unless its controller controls more lands than the attacking player")},
	})
}

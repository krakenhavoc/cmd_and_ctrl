package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Manta Ray — Creature — Fish {1}{U}{U}, 3/3:
//
//	"This creature can't attack unless defending player controls an
//	 Island.
//	 This creature can't be blocked except by blue creatures.
//	 When you control no Islands, sacrifice this creature."
//
// ADR 0107 (#1858, #1879). The attack restriction is §2's (CR 508.1c),
// with the defending player worked out per target (CR 508.5, 508.5a), so
// in Commander it may attack only the opponents who control an Island.
// The sacrifice is §1's CR 603.8 state trigger over its own controller's
// Islands: it triggers as soon as the last one is gone and not again while
// it waits or is on the stack. Both halves read one PermanentQuery.
//
// "Can't be blocked except by blue creatures" is an ordinary block rule
// (CR 509.1b) on the Ray itself.
//
// No simplification.
func init() {
	q := effects.QuerySubtype("Island")
	effects.Register(effects.Spec{
		OracleID:     "d5129531-e4b6-454e-9c67-dae925c8f2ee",
		Name:         "Manta Ray",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.CantAttackUnlessDefendingPlayerControls(q),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouControlNo(q, "Manta Ray — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnSelf(), effects.OfColor("U"), "blue creatures"),
		},
	})
}

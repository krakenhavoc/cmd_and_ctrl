package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ronom Serpent — Snow Creature — Serpent {5}{U}, 5/6:
//
//	"This creature can't attack unless defending player controls a
//	 snow land.
//	 When you control no snow lands, sacrifice this creature."
//
// ADR 0107 (#1858, #1879). The attack restriction is §2's (CR 508.1c),
// with the defending player worked out per target (CR 508.5, 508.5a), so
// in Commander it may attack only the opponents who control a snow land.
// The sacrifice is §1's CR 603.8 state trigger over its own controller's
// snow lands: it triggers as soon as the last one is gone and not again while
// it waits or is on the stack. Both halves read one PermanentQuery.
//
// No simplification.
func init() {
	q := game.PermanentQuery{Types: []string{"land"}, Supertypes: []string{"snow"}}
	Register(Spec{
		OracleID:     "ff35e480-8ea2-47bb-bb2a-24cefe9c2139",
		Name:         "Ronom Serpent",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(q),
		},
		Triggered: []game.TriggeredAbility{
			WhenYouControlNo(q, "Ronom Serpent — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}

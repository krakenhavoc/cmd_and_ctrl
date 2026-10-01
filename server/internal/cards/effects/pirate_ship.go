package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pirate Ship — Creature — Human Pirate {4}{U}, 4/3:
//
//	"This creature can't attack unless defending player controls an
//	 Island.
//	 {T}: This creature deals 1 damage to any target.
//	 When you control no Islands, sacrifice this creature."
//
// ADR 0107 (#1858, #1879). The attack restriction is §2's (CR 508.1c),
// with the defending player worked out per target (CR 508.5, 508.5a), so
// in Commander it may attack only the opponents who control an Island.
// The sacrifice is §1's CR 603.8 state trigger over its own controller's
// Islands: it triggers as soon as the last one is gone and not again while
// it waits or is on the stack. Both halves read one PermanentQuery.
//
// The tap ability is an ordinary "any target" ping, dealt by the Ship.
//
// No simplification.
func init() {
	q := QuerySubtype("Island")
	Register(Spec{
		OracleID:     "c6b3f924-806d-47d3-b044-72b48470196c",
		Name:         "Pirate Ship",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(q),
		},
		Triggered: []game.TriggeredAbility{
			WhenYouControlNo(q, "Pirate Ship — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Cost:    TapCost(),
			Targets: TargetAny(),
			Effect:  sourceDealsOneToFirstTarget,
		}},
	})
}

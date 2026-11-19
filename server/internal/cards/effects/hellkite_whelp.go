package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hellkite Whelp — Creature — Dragon {4}{R}, 3/3 (Foundations Jumpstart,
// BNS batch 32):
//
//	"Flying
//	 Whenever this creature attacks, it deals 1 damage to target creature
//	 defending player controls."
//
// Goblin Racketeer's attack trigger over TargetCreatureDefendingPlayerControls:
// the defending player is read off this attack, and with no creature there
// the trigger is removed (CR 603.3d). The damage is the Whelp's own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a9b1dd4a-2eaf-4448-a767-da705d3ec32b",
		Name:            "Hellkite Whelp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:     []game.EventKind{game.EventAttack},
			AppliesTo:   ThisAttacked,
			TargetsFrom: TargetCreatureDefendingPlayerControls,
			Key:         "Hellkite Whelp — 1 damage to target creature defending player controls",
			Effect:      sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}

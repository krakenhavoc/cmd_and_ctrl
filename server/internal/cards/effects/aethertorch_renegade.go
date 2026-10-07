package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aethertorch Renegade — Creature — Human Rogue {2}{R}, 1/2:
//
//	"When this creature enters, you get {E}{E}{E}{E} (four energy
//	 counters).
//	 {T}, Pay {E}{E}: This creature deals 1 damage to target creature.
//	 {T}, Pay eight {E}: This creature deals 6 damage to target player or
//	 planeswalker."
//
// ADR 0129 PR 1 (#1995). Both are {T} abilities of a creature, so
// summoning sickness applies (CR 302.6). The source deals the damage,
// as it last existed if it has left (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7db864fe-e0b7-4b32-8706-2ad01f93f7ca",
		Name:         "Aethertorch Renegade",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 4},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Aethertorch Renegade", 4),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{T}, Pay {E}{E}: This creature deals 1 damage to target creature.",
				Cost:    Plus(TapCost(), PayEnergy(2)),
				Targets: TargetCreature("target creature"),
				Effect:  sourceDealsDamageToEachLegalTarget(1),
			},
			{
				Label:   "{T}, Pay eight {E}: This creature deals 6 damage to target player or planeswalker.",
				Cost:    Plus(TapCost(), PayEnergy(8)),
				Targets: targetPlayerOrPlaneswalker(),
				Effect:  sourceDealsDamageToEachLegalTarget(6),
			},
		},
	})
}

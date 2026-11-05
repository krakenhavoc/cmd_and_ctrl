package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Electric Seaweed — {2}{R}{R} Creature — Plant Wall 0/4 (Teenage Mutant
// Ninja Turtles Commander, #5011):
//
//	"Defender, haste
//	 When this creature enters, until end of turn, whenever another
//	 creature dies, this creature deals 1 damage to each non-Wall
//	 creature.
//	 {T}: This creature deals 1 damage to any target."
//
// The entry trigger, as it resolves, sets up a repeating delayed trigger
// that lasts until end of turn (CR 603.7c, Shriveling Rot's shape). Each
// creature other than Electric Seaweed that dies triggers it once; the
// damage is Electric Seaweed's, from its last-known information if it has
// left (CR 608.2h), and hits every creature that isn't a Wall, so it
// spares Electric Seaweed itself. A creature that dies to that damage
// triggers it again. The tap ability is Frostwielder's ping, usable the
// turn it enters thanks to haste.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a4374baa-a846-4b34-afbf-6bb7feb4648c",
		Name:            "Electric Seaweed",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Electric Seaweed — until end of turn, whenever another creature dies, it deals 1 damage to each non-Wall creature",
				Do(DelayedOnEvent{
					Label:     "Electric Seaweed — 1 damage to each non-Wall creature",
					On:        []game.EventKind{game.EventLTB},
					Condition: tmcB2AnotherCreatureDiedCondition,
					Body:      tmcB2OneDamageToEachNonWallBody,
					Repeats:   true,
				})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Cost:    TapCost(),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}

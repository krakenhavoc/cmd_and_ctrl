package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Turntimber Basilisk — Creature — Basilisk, {1}{G}{G}, 2/1:
//
//	"Deathtouch
//	 Landfall — Whenever a land you control enters, you may have target
//	 creature block this creature this turn if able."
//
// #1684: a blocksAttacker record naming the Basilisk's own object
// (TargetBlocksThisCreature). It is usually made in a main phase,
// before the Basilisk attacks: the requirement waits for the attack and
// asks nothing if the Basilisk never attacks this turn, or if the
// target is not a creature the Basilisk's defending player controls
// (only that player's creatures can block it).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "661a99ef-8d5b-4aac-98f6-52b17a8ecf52",
		Name:            "Turntimber Basilisk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			Optional(game.TriggeredAbility{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: LandEnteredUnderYourControl,
				Targets:   TargetCreature("target creature"),
				Key:       "Turntimber Basilisk — target creature blocks this creature this turn if able",
				Effect:    TargetBlocksThisCreature,
			}, "Turntimber Basilisk — have target creature block this creature this turn if able?"),
		},
	})
}

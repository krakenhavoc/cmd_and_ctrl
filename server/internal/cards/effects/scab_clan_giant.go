package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scab-Clan Giant — Creature — Giant Warrior {4}{R}{G}, 4/5:
//
//	"When this creature enters, it fights target creature an opponent
//	 controls chosen at random."
//
// The target is ChosenAtRandom (#1815): drawn uniformly from the
// creatures your opponents control as the trigger goes on the stack,
// through the keyed random stream, and nobody is asked. With no such
// creature the trigger is removed (CR 603.3d). The fight is b10Fight,
// so a Giant that has left the battlefield by then, or a target that
// has, deals and is dealt no damage (CR 701.14b).
//
// No simplification.
func init() {
	const fight = "Scab-Clan Giant — it fights target creature an opponent controls chosen at random"
	Register(Spec{
		OracleID:     "1c2c2329-fdae-4894-a43f-41fd91cfd2a8",
		Name:         "Scab-Clan Giant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       fight,
			Targets:   ChosenAtRandom(TargetCreature("target creature an opponent controls chosen at random", OpponentControls())),
			Effect:    b30SourceFightsFirstLegalTarget,
		}},
	})
}

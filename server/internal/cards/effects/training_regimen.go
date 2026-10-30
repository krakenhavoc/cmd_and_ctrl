package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Training Regimen — {3}{G} Enchantment (Marvel Super Heroes, #4455):
//
//	"Creatures you control with +1/+1 counters on them have trample.
//	 At the beginning of combat on your turn, put a +1/+1 counter on
//	 target creature you control."
//
// The grant is a layer 6 keyword grant read live over the creatures you
// control (Tom Bombadil's KeywordGrant shape); the trigger is Avabruck
// Caretaker's beginning-of-combat target with Shortcut to Mushrooms' counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f44bde8b-43e9-4719-ac60-b8f0cee8df9b",
		Name:         "Training Regimen",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(trainingRegimenCountered, "trample"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Training Regimen — put a +1/+1 counter on target creature you control",
					PlusOneCounterOnTheTarget),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

// trainingRegimenCountered is "creatures you control with +1/+1 counters
// on them".
func trainingRegimenCountered(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller &&
		target.Counters[game.CounterPlusOne] > 0
}

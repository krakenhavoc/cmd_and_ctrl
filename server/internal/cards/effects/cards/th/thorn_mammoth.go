package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thorn Mammoth — Creature — Elephant {5}{G}{G}, 6/6 (Throne of Eldraine, S82
// old sets):
//
//	"Trample
//	 Whenever this creature or another creature you control enters, this
//	 creature fights up to one target creature you don't control."
//
// Bogwater Lumaret's one trigger on every creature entering under your control,
// itself included, and Somberwald Stag's fight against a target chosen as the
// trigger goes on the stack: "up to one" lets it trigger with no target (and do
// nothing), and a target that has gone by resolution is not fought (CR 608.2b).
// Each creature entering is its own trigger (CR 603.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d7337b13-5459-434a-aa56-e1547b875594",
		Name:            "Thorn Mammoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl,
					"Thorn Mammoth — it fights up to one target creature you don't control",
					effects.B30SourceFightsFirstLegalTarget),
				effects.TargetCreature("up to one target creature you don't control", effects.OpponentControls()).WithCount(0, 1)),
		},
	})
}

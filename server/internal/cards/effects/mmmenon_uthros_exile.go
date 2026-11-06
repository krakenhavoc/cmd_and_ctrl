package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mm'menon, Uthros Exile — Legendary Creature — Jellyfish Advisor
// {1}{U}{R}, 1/3 (Edge of Eternities, #4461):
//
//	"Flying
//	 Whenever an artifact you control enters, put a +1/+1 counter on
//	 target creature."
//
// Quicksmith Genius's watch on an artifact entering under your control
// (ArtifactEnteredUnderYourControl) with a targeted counter: the target
// is any creature, yours or an opponent's, chosen as the trigger goes on
// the stack, and a target gone by resolution gets nothing
// (PlusOneCounterOnTheTarget, CR 608.2b). Each artifact is its own
// trigger, so two Treasures entering at once put two counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6c05b882-51d9-43d8-93b8-c3317590f76a",
		Name:            "Mm'menon, Uthros Exile",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, ArtifactEnteredUnderYourControl,
				"Mm'menon, Uthros Exile — put a +1/+1 counter on target creature", PlusOneCounterOnTheTarget),
				TargetCreature("target creature")),
		},
	})
}

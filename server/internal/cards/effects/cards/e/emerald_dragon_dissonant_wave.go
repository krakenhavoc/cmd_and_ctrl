package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const emeraldDragonOracleID = "c5632f18-7904-4a29-82be-6a6ef8e85e18"

// Emerald Dragon // Dissonant Wave — Adventure.
//
//	Emerald Dragon {4}{G}{G} Creature — Dragon 4/4: "Flying, trample"
//	Dissonant Wave {2}{G} Instant — Adventure: "Counter target activated or
//	triggered ability from a noncreature source."
//
// The source is judged as it is when the ability is targeted and again as the
// spell resolves (CR 608.2b); an ability whose source has gone from every
// zone can't be shown to come from a noncreature source and is not a legal
// target. A mana ability never reaches the stack, so it can't be countered.
func init() {
	effects.Register(effects.Spec{
		OracleID:        emeraldDragonOracleID,
		Name:            "Emerald Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
	})
	effects.Register(effects.Spec{
		OracleID:     emeraldDragonOracleID + "#1",
		Name:         "Dissonant Wave",
		Completeness: effects.CompletenessFull,
		Targets: effects.AbilityOnStack("target activated or triggered ability from a noncreature source",
			effects.AbilityFromSource(effects.Noncreature())),
		OnResolve: effects.CounterTheTargetSpell,
	})
}

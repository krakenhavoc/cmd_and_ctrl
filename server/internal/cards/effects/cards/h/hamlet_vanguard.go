package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hamlet Vanguard — Creature — Human Warrior {2}{G}, 1/1 (Innistrad:
// Crimson Vow, #5495):
//
//	"Ward {2}
//	 This creature enters with two +1/+1 counters on it for each other
//	 nontoken Human you control."
//
// Ward {2} is the shared CR 702.21a trigger (Hexbane Tortoise). The entry
// is a CR 614.1c self-replacement counted as the Vanguard enters
// (Diregraf Colossus's shape): two counters per other nontoken Human its
// controller controls, read through the layers, so a changeling counts and
// a Human token does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fa0bff4f-af97-4733-b9d3-a0bc9ef1e2ca",
		Name:            "Hamlet Vanguard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"ward"},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, effects.VowB07TwoPerOtherNontokenHuman,
				"Hamlet Vanguard enters with two +1/+1 counters on it for each other nontoken Human you control"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Hamlet Vanguard — ward {2}"),
		},
	})
}

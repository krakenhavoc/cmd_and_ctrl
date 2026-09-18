package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iron Spider, Stark Upgrade — Legendary Artifact Creature — Spider
// Hero {3}, 2/3 (EDHREC rank 3195):
//
//	"Vigilance
//	 {T}: Put a +1/+1 counter on each artifact creature and/or
//	 Vehicle you control.
//	 {2}, Remove two +1/+1 counters from among artifacts you
//	 control: Draw a card."
//
// The artifact-creature deck's team pump. Vigilance rides
// PrintedKeywords; the tap ability puts a counter on every artifact
// creature and every Vehicle the controller controls, the Spider
// itself included, snapshotted before the first counter lands. A
// creature source, so the tap waits out summoning sickness (CR
// 302.6), as printed.
//
// The draw ability is live since #789, which taught
// AbilityCost.RemoveCounters to split a removal across several
// permanents: "from among artifacts you control" is
// RemoveCountersAmong, and the payment names a count per artifact
// totalling exactly two. The Spider itself is an artifact and a legal
// part of that payment — it pays for its own draw out of the counters
// its tap ability made, which is how the card is played.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e123fd7d-ace9-48a4-9510-eedcc837d8e8",
		Name:            "Iron Spider, Stark Upgrade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:  "{T}: Put a +1/+1 counter on each artifact creature and/or Vehicle you control",
			Cost:   TapCost(),
			Effect: b30PutCounterOnEachArtifactCreatureOrVehicleYouControl,
		}, {
			Label: "{2}, Remove two +1/+1 counters from among artifacts you control: Draw a card.",
			Cost: Plus(
				ManaCost("{2}"),
				RemoveCountersAmong(game.CounterPlusOne, 2, "artifacts you control", Artifact()),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}

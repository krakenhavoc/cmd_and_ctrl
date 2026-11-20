package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brinebarrow Intruder — Creature — Human Rogue {U}, 1/2 (Kaldheim,
// #5505):
//
//	"Flash
//	 When this creature enters, target creature an opponent controls
//	 gets -2/-0 until end of turn."
//
// Flash rides PrintedKeywords. The enters trigger is Binding Geist's
// shrink (vsPumpTheTarget at -2/-0) on an enters trigger instead of an
// attack trigger.
//
// No simplification.
func init() {
	const label = "Brinebarrow Intruder — target creature an opponent controls gets -2/-0 until end of turn"
	Register(Spec{
		OracleID:        "eba73313-5ec5-4004-8040-658861b71a8f",
		Name:            "Brinebarrow Intruder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
					return vsPumpTheTarget(NewContext(g, item), -2, 0, label)
				}),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

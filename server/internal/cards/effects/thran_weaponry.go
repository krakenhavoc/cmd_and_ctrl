package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thran Weaponry — Artifact, {4}:
//
//	"Echo {4} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 You may choose not to untap this artifact during your untap step.
//	 {2}, {T}: All creatures get +2/+2 for as long as this artifact remains tapped."
//
// "All creatures get +2/+2" is a continuous effect from a resolving ability,
// so the creatures it affects are locked in as it resolves (CR 611.2c), and
// it lasts for as long as this artifact remains tapped (CR 611.2b). "You may
// choose not to untap" is Amber Prison's untap opt-out.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1af303ef-dd76-4684-97a1-355b7eee53e4",
		Name:         "Thran Weaponry",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{
			mayChooseNotToUntapSelf("Thran Weaponry — you may choose not to untap this artifact"),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: All creatures get +2/+2 for as long as this artifact remains tapped.",
			Cost:  Plus(ManaCost("{2}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				d, ok := DurationWhileThisRemainsTapped(ctx)
				if !ok {
					return nil
				}
				return ScopedEffectFor{
					Match:    Creature(),
					Mods:     []game.Mod{game.ModifyPTMod(2, 2)},
					Duration: d,
					Label:    "Thran Weaponry — +2/+2 while tapped",
				}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{Echo("Thran Weaponry", "{4}")},
	})
}

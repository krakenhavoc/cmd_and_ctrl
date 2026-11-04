package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Festering Newt — Creature — Salamander {B}, 1/1:
//
//	"When this creature dies, target creature an opponent controls gets
//	 -1/-1 until end of turn. That creature gets -4/-4 instead if you
//	 control a creature named Bogbrew Witch."
//
// One trigger. "If you control a creature named Bogbrew Witch" is not an
// intervening-if (it does not follow "when"), so the trigger always goes
// on the stack and the Witch is looked for as it resolves: a Witch that
// arrives in response makes it -4/-4, one that leaves makes it -1/-1.
// "You" is the trigger's controller, who controlled the Newt as it died.
// A target that has left or changed control by resolution makes the
// trigger do nothing (CR 608.2b).
//
// No simplification.
func init() {
	t := WhenThisDies("Festering Newt — target creature an opponent controls gets -1/-1 (or -4/-4 with Bogbrew Witch)",
		func(g *game.Game, item *game.StackItem) error {
			if len(item.Targets) == 0 {
				return nil
			}
			n := 1
			if jmpB1YouControlACreatureNamed(g, item.Controller, "Bogbrew Witch") {
				n = 4
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -n,
				Toughness: -n,
				Label:     "Festering Newt — -1/-1 or -4/-4 until end of turn",
			}.Apply(NewContext(g, item))
		})
	t.Targets = TargetCreature("target creature an opponent controls", OpponentControls())
	Register(Spec{
		OracleID:     "5887922e-d1fb-44e2-9498-2065ce8ac870",
		Name:         "Festering Newt",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

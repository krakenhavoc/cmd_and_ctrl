package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Erhnam Djinn — Creature — Djinn (4/5) for {3}{G}:
//
//	"At the beginning of your upkeep, target non-Wall creature an opponent
//	 controls gains forestwalk until your next upkeep. (It can't be blocked
//	 as long as defending player controls a Forest.)"
//
// The grant is a layer-6 forestwalk on the target, locked to that object
// (CR 400.7), and timed by game.UntilYourNextUpkeep (#5737): it ends as
// the trigger's controller's next upkeep BEGINS, before that upkeep's own
// Erhnam trigger is put on the stack (CR 503.1a), so the new target is
// chosen on a board where last turn's forestwalk is already gone. "Your"
// is the trigger's controller. The grant outlives the Djinn, and if its
// controller leaves the game it ends as their next turn would have begun
// (CR 800.4m).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d48a38c9-3dcd-4c18-8840-1b057ede3ff0",
		Name:         "Erhnam Djinn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourUpkeep("Erhnam Djinn — target non-Wall creature an opponent controls gains forestwalk until your next upkeep",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					target := FirstLegalBattlefieldTarget(ctx)
					if target == uuid.Nil {
						return nil
					}
					return ScopedEffectFor{
						Target:   target,
						Mods:     []game.Mod{game.AddKeywordsMod("forestwalk")},
						Duration: DurationUntilYourNextUpkeep(ctx, item.Controller),
						Label:    "Erhnam Djinn — forestwalk until your next upkeep",
					}.Apply(ctx)
				}), TargetCreature("target non-Wall creature an opponent controls", NonWall(), OpponentControls())),
		},
	})
}

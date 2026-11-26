package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kinsbaile Skirmisher — Creature — Kithkin Soldier {1}{W}, 2/2 (Magic 2015,
// #5556):
//
//	"When this creature enters, target creature gets +1/+1 until end of turn."
//
// The target may be this creature itself, and is chosen as the trigger goes on
// the stack; one that has left or become illegal by resolution is not pumped
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b142a4a-769b-4e1b-aae6-ba6e11f35836",
		Name:         "Kinsbaile Skirmisher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Kinsbaile Skirmisher — target creature gets +1/+1 until end of turn",
				oldB356SkirmisherPump), effects.TargetCreature("target creature")),
		},
	})
}

func oldB356SkirmisherPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return effects.BoostUntilEOT{Target: targets[0].ID, Power: 1, Toughness: 1, Label: "Kinsbaile Skirmisher — +1/+1"}.Apply(ctx)
}

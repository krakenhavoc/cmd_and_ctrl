package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Explosive Prodigy — Creature — Elemental Sorcerer {1}{R}, 1/1 (Lorwyn
// Eclipsed, #4458):
//
//	"Vivid — When this creature enters, it deals X damage to target
//	 creature an opponent controls, where X is the number of colors among
//	 permanents you control."
//
// Vivid is an ability word with no rules meaning. X is Mondo Gecko's
// colour count (colorsAmongPermanentsYouControl), read as the trigger
// resolves, so a permanent that left in response no longer adds its
// colours. The target is required: with no creature an opponent controls,
// the trigger is removed. The damage is the Prodigy's, read as it last
// existed if it has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "22d059fc-4e47-44b5-9e21-4a56f1dff6a7",
		Name:         "Explosive Prodigy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Explosive Prodigy — X damage to target creature an opponent controls (vivid)",
					explosiveProdigyDamage),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

// explosiveProdigyDamage deals X, the colours among the controller's
// permanents now, to the announced creature if it is still legal.
func explosiveProdigyDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := len(colorsAmongPermanentsYouControl(g, item.Controller))
	if x == 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: x}.Apply(ctx)
	}
	return nil
}

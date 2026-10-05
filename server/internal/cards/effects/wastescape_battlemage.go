package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wastescape Battlemage — Creature — Eldrazi Wizard {1}{C}, 2/2:
//
//	"Kicker {G} and/or {1}{U}
//	 When you cast this spell, if it was kicked with its {G} kicker,
//	 exile target artifact or enchantment an opponent controls.
//	 When you cast this spell, if it was kicked with its {1}{U} kicker,
//	 return target creature an opponent controls to its owner's hand."
//
// The one Battlemage whose linked abilities are CAST triggers (CR
// 702.33f, CR 603.4) rather than enters triggers: they go on the stack
// above the spell, so they resolve even if the Battlemage is countered.
// The kicker record is read off the spell's stack item, where an enters
// trigger would read the permanent's (castTriggerIfKickedWith).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "3cc6475d-aae9-4592-a26b-0fbb4269759d",
		Name:          "Wastescape Battlemage",
		Completeness:  CompletenessFull,
		OptionalCosts: Kickers("{G}", "{1}{U}"),
		Triggered: []game.TriggeredAbility{
			castTriggerIfKickedWith("{G}",
				"Wastescape Battlemage — kicked with {G}, exile target artifact or enchantment an opponent controls",
				TargetPermanent("target artifact or enchantment an opponent controls", And(Or(Artifact(), Enchantment()), OpponentControls())),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return ExileTarget{Target: t.ID}.Apply(ctx)
					}
					return nil
				}),
			castTriggerIfKickedWith("{1}{U}",
				"Wastescape Battlemage — kicked with {1}{U}, return target creature an opponent controls to its owner's hand",
				TargetCreature("target creature an opponent controls", OpponentControls()),
				func(g *game.Game, item *game.StackItem) error { return bounceTheTarget(item, NewContext(g, item)) }),
		},
	})
}

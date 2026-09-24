package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Sharpshooter —
//
// "This creature doesn't untap during your untap step. Whenever a creature
// dies, untap this creature. {T}: This creature deals 1 damage to any target."
func init() {
	Register(Spec{
		OracleID:              "d81285b7-a718-411a-8be3-ecc0cfe0bcb0",
		Name:                  "Goblin Sharpshooter",
		Completeness:          CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Goblin Sharpshooter deals 1 damage to any target",
			Cost:    TapCost(),
			Targets: TargetAny(),
			Effect:  sourceDealsOneToFirstTarget,
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := diedCreature(ev, g)
				return died
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Goblin Sharpshooter — untap", func(g *game.Game, item *game.StackItem) error {
					return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
				})
			}}},
	})
}

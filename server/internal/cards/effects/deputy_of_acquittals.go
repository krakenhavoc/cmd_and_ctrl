package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deputy of Acquittals — "Flash. When this creature enters, you may
// return another target creature you control to its owner's hand."
//
// Sandbox simplification: the target clause cannot express
// "another" — TargetSpec is declared statically at init() and
// NotSelf needs the source's InstanceID, which only exists once the
// card is on the battlefield. The picker therefore offers every
// creature the controller controls, Deputy included; the Effect
// declines to bounce itself so the printed restriction still holds
// at resolution.
func init() {
	Register(Spec{
		OracleID:        "3cbb5045-8566-4279-b7d3-3e599b11ccc5",
		Name:            "Deputy of Acquittals",
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetCreature("target creature you control", YouControl()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Deputy of Acquittals — return a creature to hand",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 {
							return nil
						}
						target := item.Targets[0]
						if target.ID == item.SourceCardID {
							return nil
						}
						return BounceToHand{Target: target.ID}.Apply(NewContext(g, item))
					})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Deputy of Acquittals — return a creature you control to its owner's hand?",
			},
		}},
	})
}

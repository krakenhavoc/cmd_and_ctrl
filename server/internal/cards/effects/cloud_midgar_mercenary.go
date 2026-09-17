package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloud, Midgar Mercenary — "When Cloud enters, search your library for an
// Equipment card, reveal it, put it into your hand, then shuffle. As long as
// Cloud is equipped, if a triggered ability of Cloud or an Equipment attached
// to it triggers, that ability triggers an additional time."
//
// The trigger harvester retains the pre-move attachment relationship for a
// leaving Equipment, so both halves of the card are covered.
func init() {
	Register(Spec{
		OracleID:        "33d2584b-bf29-4c22-bd45-14ba2fb98c0e",
		Name:            "Cloud, Midgar Mercenary",
		Completeness:    CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{cloudTriggerDoubler()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cloud, Midgar Mercenary — search for an Equipment card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: b09IsEquipmentCard,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Cloud, Midgar Mercenary — an Equipment card, revealed, to hand",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

func cloudTriggerDoubler() game.TriggerDoubler {
	d := DoublesAbilitiesOf(nil, DoublesAbilitiesOfOptions{
		SkipControllerCheck: true,
		SourceMatch: func(g *game.Game, q game.TriggerDoublingQuery) bool {
			if q.Source.InstanceID == q.Doubler.InstanceID {
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.HasSubtype("Equipment") && c.IsAttachedTo(q.Doubler.InstanceID) {
						return true
					}
				}
				return false
			}
			return q.Source.HasSubtype("Equipment") && q.Source.IsAttachedTo(q.Doubler.InstanceID)
		},
	})
	d.Label = "Cloud, Midgar Mercenary"
	return d
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sangromancer — Creature — Vampire Shaman {2}{B}{B}, 3/3 (EDHREC
// rank 2008):
//
//	"Flying
//	 Whenever a creature an opponent controls dies, you may gain 3
//	 life.
//	 Whenever an opponent discards a card, you may gain 3 life."
//
// Two optional lifegain triggers, one per printed ability: the
// dead creature is read post-move (its controller survives the
// move), and the discarding player is the discard event's Actor.
// A wheel against three opponents is one prompt per card, as
// printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "920445ab-0ac2-4de7-bc1c-f5e58eb4424c",
		Name:            "Sangromancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b18OpponentsCreatureDied(ev, source, g)
				},
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Sangromancer — an opponent's creature died. Gain 3 life?"},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Sangromancer — gain 3 life (a creature died)", b18GainThree)
				},
			},
			{
				Watches: []game.EventKind{game.EventDiscardCard},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return b18OpponentDiscarded(ev, source)
				},
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Sangromancer — an opponent discarded. Gain 3 life?"},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Sangromancer — gain 3 life (a card was discarded)", b18GainThree)
				},
			},
		},
	})
}

// b18GainThree is Sangromancer's shared effect body.
func b18GainThree(g *game.Game, item *game.StackItem) error {
	return GainLife{Player: item.Controller, Amount: 3}.Apply(NewContext(g, item))
}

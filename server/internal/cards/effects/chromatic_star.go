package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chromatic Star — Artifact {1} (EDHREC rank 2832):
//
//	"{1}, {T}, Sacrifice this artifact: Add one mana of any color.
//	 When this artifact is put into a graveyard from the battlefield,
//	 draw a card."
//
// The colour-fixing cantrip egg. The mana ability carries all three
// cost components — {1} paid from the pool, the tap, and the
// sacrifice — validated together so a failed activation leaves the
// Star untouched; the any-colour pick keeps its printed width. The
// draw is a leaves-for-the-graveyard trigger on the Star itself, not
// a rider on the ability: it fires whether the Star was cracked,
// destroyed or sacrificed to something else, as printed, and it uses
// the stack (the mana ability does not), so the mana is in the pool
// before the card is drawn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fedbd40b-e3a5-449c-a8a3-b42e9da191a9",
		Name:         "Chromatic Star",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true, Sacrifice: true, Mana: "{1}"},
			Produced:                "{W|U|B|R|G}",
			Label:                   "{1}, {T}, Sacrifice: Add one mana of any color",
			IgnoreCommanderIdentity: true,
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Chromatic Star — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}

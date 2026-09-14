package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Khalni Garden — Land (EDHREC rank 3175):
//
//	"This land enters tapped.
//	 When this land enters, create a 0/1 green Plant creature token.
//	 {T}: Add {G}."
//
// A Forest that brings a chump blocker. The tapped entry is the real
// CR 614 self-replacement; the Plant is an enters trigger on the
// stack, so it can be responded to; the mana ability is a plain
// green tap. The Plant is Avenger of Zendikar's 0/1 green Plant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b2d5ba45-8674-4428-89db-c2bbbf0bf5c5",
		Name:         "Khalni Garden",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Khalni Garden — create a 0/1 green Plant",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{Controller: item.Controller, Template: b02PlantToken(), N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Lion-Turtle — Legendary Creature — Elder Cat Turtle {1}{G}{U}, 3/6
// (Avatar: The Last Airbender, #4459):
//
//	"Reach, vigilance
//	 When The Lion-Turtle enters, you gain 3 life.
//	 The Lion-Turtle can't attack or block unless there are three or more
//	 Lesson cards in your graveyard.
//	 {T}: Add one mana of any color."
//
// The restriction is Hazoret's shape: a layer 6 static on itself that is
// on only while its controller's graveyard holds fewer than three Lesson
// cards (avB4LessonsInGraveyard), setting CantAttack and CantBlock, read
// live as attackers and blockers are declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fff728a1-56f5-4e97-91e2-19aba5772909",
		Name:            "The Lion-Turtle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		Static: []game.StaticAbility{{
			Label: "Can't attack or block unless there are three or more Lesson cards in your graveyard.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfOnly(target, g, source) && avB4LessonsInGraveyard(g, source.Controller) < 3
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("The Lion-Turtle — you gain 3 life", Do(GainLife{Amount: 3})),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}

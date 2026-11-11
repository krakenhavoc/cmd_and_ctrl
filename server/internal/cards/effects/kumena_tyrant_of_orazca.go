package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kumena, Tyrant of Orazca — Legendary Creature — Merfolk Shaman {1}{G}{U}, 2/4:
//
//	"Tap another untapped Merfolk you control: Kumena can't be blocked
//	 this turn.
//	 Tap three untapped Merfolk you control: Draw a card.
//	 Tap five untapped Merfolk you control: Put a +1/+1 counter on each
//	 Merfolk you control."
//
// All three are tap-others costs (CR 602.2): tapping them is not the {T}
// symbol, so a Merfolk with summoning sickness may pay (CR 302.6). Only
// the first says "another"; the other two may tap Kumena itself, since it
// is an untapped Merfolk you control. Each ability can be activated again
// as long as enough untapped Merfolk remain.
func init() {
	Register(Spec{
		OracleID:     "e7125493-febc-4e8e-abdc-5fc4f94a281c",
		Name:         "Kumena, Tyrant of Orazca",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "Tap another untapped Merfolk you control: Kumena can't be blocked this turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    cmdr2B05TapMerfolk(1, true, "another untapped Merfolk you control"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
						Label: "Kumena, Tyrant of Orazca — can't be blocked this turn"}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "Tap three untapped Merfolk you control: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    cmdr2B05TapMerfolk(3, false, "three untapped Merfolk you control"),
				Effect:  Do(DrawCards{N: 1}),
			},
			{
				Label:   "Tap five untapped Merfolk you control: Put a +1/+1 counter on each Merfolk you control.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    cmdr2B05TapMerfolk(5, false, "five untapped Merfolk you control"),
				Effect:  cmdr2B05CountersOnEachMerfolk(1),
			},
		},
	})
}

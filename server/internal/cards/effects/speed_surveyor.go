package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// maxSpeedSurveyorDraw is the ability the five Surveyors share:
//
//	"Max speed — {3}, Exile this card from your graveyard: Draw a card."
//
// ADR 0138 (#2122, #2711). It functions from the graveyard (CR 113.6,
// ADR 0062), and so does the max speed ability that grants it
// (CR 702.178b), so it works there while its owner's speed is 4. One
// builder rather than five copies of the body (TestNoNewExactClonesInTheCatalog).
func maxSpeedSurveyorDraw() ActivatedAbility {
	return MaxSpeedActivated(ActivatedAbility{
		Label:   "Max speed — {3}, Exile this card from your graveyard: Draw a card.",
		Purpose: game.Purpose{Answers: game.AnswerValue},
		Cost:    Plus(ManaCost("{3}"), ExileThis()),
		Zones:   []game.ZoneKind{game.ZoneGraveyard},
		Effect: func(g *game.Game, item *game.StackItem) error {
			return DrawCards{N: 1}.Apply(NewContext(g, item))
		},
	})
}

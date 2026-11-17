package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clay Revenant — Artifact Creature — Golem {1}, 1/2 (The Brothers' War, #5482):
//
//	"This creature enters tapped.
//	 {2}{B}: Return this card from your graveyard to your hand."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c183186b-d92c-45a2-92d1-42e0df2b441f",
		Name:         "Clay Revenant",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{2}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  returnThisCardFromYourGraveyardToYourHand,
		}},
	})
}

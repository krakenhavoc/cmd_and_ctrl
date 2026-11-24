package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prismari Pledgemage — Creature — Orc Wizard {U/R}{U/R}, 3/3 (Strixhaven:
// School of Mages, #5502):
//
//	"Defender
//	 Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	 this creature can attack this turn as though it didn't have defender."
//
// Defender is the printed keyword. The magecraft is Hightide Hermit's
// until-end-of-turn defender waiver (CR 702.3b, CR 609.4) on the source,
// which keeps the keyword and does nothing for a creature that has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "47ecedec-7bf7-4a3b-ac33-2bf203fa50d9",
		Name:            "Prismari Pledgemage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.Magecraft("Prismari Pledgemage — can attack this turn as though it didn't have defender (magecraft)",
				effects.CanAttackThisTurnAsThoughNoDefender("Prismari Pledgemage — can attack this turn as though it didn't have defender")),
		},
	})
}

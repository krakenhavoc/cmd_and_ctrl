package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valiant Changeling — Creature — Shapeshifter {5}{W}{W}, 3/3 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"This spell costs {1} less to cast for each creature type among
//	 creatures you control. This effect can't reduce the amount of mana this
//	 spell costs by more than {5}.
//	 Changeling (This card is every creature type.)
//	 Double strike"
//
// Bedlam Reveler's self cost modifier, over the distinct creature types
// among the caster's creatures (each type once however many creatures share
// it, CR 205.3m; a changeling is every creature type, CR 702.73a, so it
// alone reaches the cap), read through the layers, and capped at five. The
// cap is the effect's own: the amount is min(types, 5). Changeling and
// double strike are printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "65d9b8b6-5215-4037-8e50-e0ca80c5dc9e",
		Name:            "Valiant Changeling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"changeling", "double strike"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.CreatureTypesAmongYourCreaturesMax5,
				"This spell costs {1} less to cast for each creature type among creatures you control. This effect can't reduce the amount of mana this spell costs by more than {5}."),
		},
	})
}

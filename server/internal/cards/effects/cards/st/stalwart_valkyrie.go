package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stalwart Valkyrie — {3}{W} Creature — Angel Warrior 3/2 (Foundations
// Jumpstart):
//
//	"You may pay {1}{W} and exile a creature card from your graveyard
//	 rather than pay this spell's mana cost.
//	 Flying"
//
// An alternative cost (CR 118.9) with a mana half and an exile-from-your-
// graveyard half, named by the caster in alt_cost_ids and exiled with the
// spell on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b6b4c459-9c8b-4b95-8706-05d972b8daea",
		Name:            "Stalwart Valkyrie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{{
			Key:                "exile-creature-from-graveyard",
			Label:              "Pay {1}{W} and exile a creature card from your graveyard rather than pay this spell's mana cost",
			ManaCost:           "{1}{W}",
			ExileFromGraveyard: effects.CardsInYourGraveyard(1, "a creature card", effects.Creature()),
			PayLabel:           "a creature card from your graveyard",
		}},
	})
}

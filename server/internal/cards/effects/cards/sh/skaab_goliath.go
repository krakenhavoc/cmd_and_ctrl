package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skaab Goliath — Creature — Zombie Giant {5}{U}, 6/9 (Commander Legends):
//
//	"As an additional cost to cast this spell, exile two creature cards from
//	 your graveyard.
//	 Trample"
//
// Skaab Ruinator's cost: the mandatory exile-cards slot of a cast, two
// creature cards in the caster's own graveyard, paid with the spell on the
// stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "498707db-258f-423a-9ff6-9e294fce84d6",
		Name:            "Skaab Goliath",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		AdditionalCost:  effects.ExileFromGraveyardCost(2, "two creature cards", func(c game.Card) bool { return c.IsCreature() }),
	})
}

package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invigorating Falls — Sorcery {2}{G}{G} (Torment, #5612):
//
//	"You gain life equal to the number of creature cards in all
//	 graveyards."
//
// Blossoming Wreath over every graveyard: the creature cards are counted
// once, as the spell resolves (CR 608.2h), with Mortivore's count
// (b41CreatureCardsInAllGraveyards). No creature cards is no life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5c98759-e009-4340-9326-9f5ff70a86bc",
		Name:         "Invigorating Falls",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GainLife{Amount: effects.B41CreatureCardsInAllGraveyards(ctx.Game)}.Apply(ctx)
		},
	})
}

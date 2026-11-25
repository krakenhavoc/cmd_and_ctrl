package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Witch's Vengeance — Sorcery {1}{B}{B} (Throne of Eldraine, #5516):
//
//	"Creatures of the creature type of your choice get -3/-3 until end of
//	 turn."
//
// Tribal Unity's choose-a-type prompt; the creatures are fixed when the
// spell resolves (CR 611.2c) and include every player's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b0d9139-b3ad-4065-9b55-aeeda8950bee",
		Name:         "Witch's Vengeance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 3, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			effects.ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID, "Witch's Vengeance — choose a creature type",
				ctx.ItemTail(witchsVengeanceTypeChosen))
			return nil
		},
	})
}

var witchsVengeanceTypeChosen = effects.CreatureTypeChosen("then/witchs-vengeance-type-chosen", func(ctx *effects.Context, t string, _ game.TailOutcome) error {
	return effects.BoostUntilEOT{
		Match:     effects.OfCreatureType(t),
		Power:     -3,
		Toughness: -3,
		Label:     "Witch's Vengeance — -3/-3",
	}.Apply(ctx)
})

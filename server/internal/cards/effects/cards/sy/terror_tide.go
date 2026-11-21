package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Terror Tide — Sorcery {2}{B}{B} (The Lost Caverns of Ixalan, #4470):
//
//	"Fathomless descent — All creatures get -X/-X until end of turn,
//	 where X is the number of permanent cards in your graveyard."
//
// X is counted as the spell resolves, with the spell itself already on
// its way to the graveyard, which is not a permanent card (CR 110.4), and
// the creatures it affects are fixed at that moment (CR 611.2c): a creature
// that arrives later this turn is not shrunk. A card cast from the
// graveyard or a creature the shrink kills adds nothing to X, which was
// settled first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12be58f4-1c12-4bca-8069-1f049cd8a90c",
		Name:         "Terror Tide",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := effects.PermanentCardsInGraveyard(ctx.Game, ctx.Controller())
			if x <= 0 {
				return nil
			}
			return effects.BoostUntilEOT{Match: effects.Creature(), Power: -x, Toughness: -x}.Apply(ctx)
		},
	})
}

package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blight Grenade — Sorcery {4}{B} (Warhammer 40,000 Commander, #5948):
//
//	"Destroy target creature.
//	 All creatures get -3/-3 until end of turn."
//
// The target is destroyed first; then every creature on the battlefield as
// the spell resolves gets -3/-3 (CR 611.2c), a creature that enters later
// does not. A target that became illegal fizzles the whole spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae24b863-974c-4efa-ab41-23cfa1906c54",
		Name:         "Blight Grenade",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose: game.Purpose{
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 3},
			Answers: game.AnswerRemove,
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return effects.BoostUntilEOT{Match: effects.Creature(), Power: -3, Toughness: -3,
				Label: "Blight Grenade — all creatures get -3/-3"}.Apply(ctx)
		},
	})
}

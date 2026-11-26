package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// In Oketra's Name — {1}{W} Instant (Amonkhet, S82 old batch 246):
//
//	"Zombies you control get +2/+1 until end of turn. Other creatures you
//	 control get +1/+1 until end of turn."
//
// Both sets are chosen as the spell resolves (CR 611.2c): a Zombie gets only
// the first bonus, every other creature the second. A Changeling or a type
// change is read through the layered view.
//
// No simplification.
func init() {
	zombie := effects.HasSubtype("Zombie")
	effects.Register(effects.Spec{
		OracleID:     "f4a8756d-94df-40bc-a363-2938ada0a808",
		Name:         "In Oketra's Name",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl(), zombie), Power: 2, Toughness: 1,
				Label: "In Oketra's Name — Zombies +2/+1",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl(), effects.Not(zombie)), Power: 1, Toughness: 1,
				Label: "In Oketra's Name — other creatures +1/+1",
			}.Apply(ctx)
		},
	})
}

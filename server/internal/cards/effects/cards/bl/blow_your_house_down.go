package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blow Your House Down — Sorcery {2}{R} (Throne of Eldraine, S82 old sets):
//
//	"Up to three target creatures can't block this turn. Destroy any of them
//	 that are Walls."
//
// Each still-legal target gets the can't-block restriction until end of turn
// (CR 509.1b), then each that is a Wall as the spell resolves is destroyed
// (CR 701.8). The subtype is read through the layers, so a creature that was
// turned into a Wall is destroyed and a Wall that stopped being one is not.
//
// No simplification.
func init() {
	const label = "Blow Your House Down — can't block this turn"
	effects.Register(effects.Spec{
		OracleID:     "a44e8d98-3bc8-42ce-b09f-a3b947d9ce53",
		Name:         "Blow Your House Down",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to three target creatures").WithCount(0, 3),
		Purpose:      effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			for _, t := range legal {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: label}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, t := range legal {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok || !c.HasSubtype(ctx.TextSubtype("Wall")) {
					continue
				}
				if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

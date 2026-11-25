package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wrap in Flames — Sorcery {3}{R} (Theros Beyond Death, #5514):
//
//	"Wrap in Flames deals 1 damage to each of up to three target creatures. Those
//	 creatures can't block this turn."
//
// One damage event from the spell to every target still legal as it resolves
// (CR 608.2b), then each of those creatures can't block until end of turn (CR
// 611.2a), a creature that died to the damage simply being gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b46efd59-eda7-4fe6-ace7-67ca8f6c088d",
		Name:         "Wrap in Flames",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to three target creatures").WithCount(0, 3),
		Purpose: effects.ForTargets(game.TargetPurpose{
			Slot: 0, Damage: 1, UntilEndOfTurn: game.TargetHarm, Pump: &game.Pump{CantBlock: true},
		}),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := effects.B10DamageEachLegalTarget(ctx, 1); err != nil {
				return err
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: "Wrap in Flames — can't block"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

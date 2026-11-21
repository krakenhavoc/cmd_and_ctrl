package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lacerate Flesh — Sorcery {4}{R} (Innistrad: Crimson Vow, #5495):
//
//	"Lacerate Flesh deals 4 damage to target creature. Create a number of
//	 Blood tokens equal to the amount of excess damage dealt to that
//	 creature this way."
//
// Archaic's Agony's reading of excess damage (CR 120.4a: what landed beyond
// the lethal damage the creature still needed, measured off the damage event
// so a prevention shield lowers it too). Each point of excess is a Blood
// token for the caster; with no excess nothing is created.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "201c8de7-9675-475b-8f9e-7b1935234488",
		Name:         "Lacerate Flesh",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				excess, err := effects.SocB20DealDamageWithExcess(ctx, t.ID, 4)
				if err != nil || excess <= 0 {
					return err
				}
				return effects.CreateToken{Controller: item.Controller, Template: effects.BloodToken(), N: excess}.Apply(ctx)
			}
			return nil
		},
	})
}

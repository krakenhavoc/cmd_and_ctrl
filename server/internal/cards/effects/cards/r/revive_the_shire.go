package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Revive the Shire — Sorcery {1}{G} (EDHREC rank 2950):
//
//	"Return target permanent card from your graveyard to your hand.
//	 Create a Food token. (It's an artifact with "{2}, {T}, Sacrifice
//	 this token: You gain 3 life.")"
//
// Regrowth for permanents with a Food on the side. The target clause
// is "permanent card in your graveyard" (Permanent() over the
// graveyard zone); a target that left in response counters the spell
// by game rules, Food included — CR 608.2b, as printed. The Food is
// the shared FoodToken template.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67afcfef-3f0c-4823-b690-01732f28a005",
		Name:         "Revive the Shire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target permanent card from your graveyard", effects.YouOwn(), effects.Permanent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.FoodToken(), N: 1}.Apply(ctx)
		},
	})
}

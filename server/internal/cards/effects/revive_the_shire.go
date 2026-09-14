package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "67afcfef-3f0c-4823-b690-01732f28a005",
		Name:         "Revive the Shire",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target permanent card from your graveyard", YouOwn(), Permanent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{Controller: ctx.Controller(), Template: FoodToken(), N: 1}.Apply(ctx)
		},
	})
}

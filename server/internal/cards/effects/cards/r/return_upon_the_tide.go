package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Return Upon the Tide — Sorcery {4}{B} (Kaldheim):
//
//	"Return target creature card from your graveyard to the battlefield. If
//	 it's an Elf, create two 1/1 green Elf Warrior creature tokens.
//	 Foretell {3}{B} (During your turn, you may pay {2} and exile this card
//	 from your hand face down. Cast it on a later turn for its foretell cost.)"
//
// The card is read as it entered (the creature types it has on the
// battlefield, so a changeling or a granted type counts) once it is there; a
// card that failed to return gives no tokens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "b0bd6135-4257-4f71-bdfe-71fb3f7a36a5",
		Name:           "Return Upon the Tide",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Foretell("{3}{B}")},
		Targets:        effects.TargetCardInGraveyard("target creature card in your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ids := effects.LegalTargetCardIDs(ctx)
			if len(ids) == 0 {
				return nil
			}
			if err := effects.ReturnFirstLegalGraveyardTargetToBattlefield(ctx.Game, item); err != nil {
				return err
			}
			if z := ctx.Game.FindCardZoneForEffect(ids[0]); z == nil || z.Kind != game.ZoneBattlefield {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(ids[0])
			if !ok || !c.HasSubtype(ctx.TextSubtype("Elf")) {
				return nil
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: 2}.Apply(ctx)
		},
	})
}

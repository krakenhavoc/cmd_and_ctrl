package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reunion of the House — Sorcery {5}{W}{W} (Tarkir: Dragonstorm Commander):
//
//	"Return any number of target creature cards with total power 10 or less
//	 from your graveyard to the battlefield. Exile Reunion of the House."
//
// The clause is "any number of target creature cards" in your graveyard
// with a TargetTotal over power (game.TargetTotalPower, beside Fire Lord
// Sozin's mana-value total): the announce gate sums the picks, a pick set
// whose power has grown past ten by resolution is wholly illegal (the
// ruling), and power-0 cards always fit. Power is read where the card is,
// the graveyard. Every target still legal returns together as one entry
// (Fire Lord Sozin's door), then the spell is exiled instead of going to
// the graveyard, whether or not anything returned.
//
// Seam: game.TargetTotalPower.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad5ff61f-79d6-4dd9-b4fe-17f7a0055c44",
		Name:         "Reunion of the House",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("any number of target creature cards with total power 10 or less from your graveyard",
			effects.Creature(), effects.YouOwn()).WithCount(0, 0).WithTotal(effects.CmdrB13TotalPowerAtMost(10)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				var ids []uuid.UUID
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						ids = append(ids, t.ID)
					}
				}
				if len(ids) == 0 {
					return nil
				}
				return effects.PutCardsFromZoneTogether(ctx, game.ZoneGraveyard, ids, ctx.Controller(), false)
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.ExileTarget{Target: ctx.Source()}.Apply(ctx) }),
		),
	})
}

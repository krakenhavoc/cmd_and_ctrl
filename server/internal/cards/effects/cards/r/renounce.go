package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Renounce — Instant {1}{W} (Mercadian Masques, #5619):
//
//	"Sacrifice any number of permanents. You gain 2 life for each
//	 permanent sacrificed this way."
//
// The controller picks any number of the permanents they control, zero
// included, as the spell resolves; they are sacrificed together and the
// life is 2 for each one that was really sacrificed (a keyed tail told the
// permanents that left, CR 701.21).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "72f9a6f7-ea26-438d-aafa-758239fd3681",
		Name:         "Renounce",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() {
					ids = append(ids, c.InstanceID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  ctx.Controller(),
				Source:   ctx.Source(),
				Question: "Renounce — sacrifice any number of permanents (you gain 2 life for each)",
				Cards:    ids,
				Min:      0,
				Max:      len(ids),
				LiveZone: game.ZoneBattlefield,
				Asked:    game.NewAskedItem(item),
				ThenKey:  renounceSacrifice,
			})
			return nil
		},
	})
}

// renounceSacrifice sacrifices the chosen permanents together and hands the
// ones that were sacrificed to the life tail. An on-disk key.
var renounceSacrifice = effects.AskedPick("choose-cards/s82-old-b694-renounce-sacrifice", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ctx.Game.SacrificeAllKeyedForEffect(ctx.Source(), r.Picked, ctx.ItemTail(renounceGain))
})

// renounceGain is "you gain 2 life for each permanent sacrificed this way".
// An on-disk key.
var renounceGain = effects.CardTail("s82-old-b694-renounce/gain", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.GainLife{Player: ctx.Controller(), Amount: 2 * len(o.IDs)}.Apply(ctx)
})

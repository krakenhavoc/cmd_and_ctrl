package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wild Mammoth — Creature — Elephant {2}{G}, 3/4 (Nemesis, S82 OLD batch
// 680, #5618):
//
//	"At the beginning of your upkeep, if a player controls more creatures
//	 than each other player, the player who controls the most creatures
//	 gains control of this creature."
//
// An intervening "if" (CR 603.4) read at the trigger and again as it
// resolves, over creatures (tokens included) as the layers see them. A tie
// for the most names nobody. The player who gains control is read as the
// ability resolves; when that is the Mammoth's own controller nothing
// changes. The control is until the Mammoth leaves or another effect
// changes it (CR 611.2a), so it can move again at the new controller's
// upkeep.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29ef65bf-1066-44a5-a2eb-e3670b076647",
		Name:         "Wild Mammoth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("a player controls more creatures than each other player",
				func(g *game.Game, _ uuid.UUID) bool { return effects.OldB680MostCreatures(g) != uuid.Nil },
				effects.AtYourUpkeep("Wild Mammoth — the player who controls the most creatures gains control of this creature",
					effects.Do(effects.Instruction(func(ctx *effects.Context) error {
						most := effects.OldB680MostCreatures(ctx.Game)
						if most == uuid.Nil || most == ctx.Controller() || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
							return nil
						}
						return effects.GainControl{
							Target:     ctx.Source(),
							Controller: most,
							Duration:   game.IndefiniteDuration(),
							Label:      "Wild Mammoth — the player who controls the most creatures gains control of it",
						}.Apply(ctx)
					})))),
		},
	})
}

package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Felidar Retreat — Enchantment {3}{W} (EDHREC top 1,000, #3654):
//
//	"Landfall — Whenever a land you control enters, choose one —
//	 • Create a 2/2 white Cat Beast creature token.
//	 • Put a +1/+1 counter on each creature you control. Those
//	   creatures gain vigilance until end of turn."
//
// A landfall trigger with a mode choice, built as Venser, Fervent
// Forger builds its ETB: the trigger itself does nothing and each
// bullet is a ModeDoing body, so the choice is the CR 603.3c
// mode_pick prompt. Neither bullet targets. "Those creatures" in the
// second bullet are the ones that got a counter, fixed as the bullet
// resolves.
//
// No simplification.
func init() {
	landfall := effects.Landfall("Felidar Retreat — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	landfall.Modes = effects.ChooseOne(
		effects.ModeDoing("Create a 2/2 white Cat Beast creature token.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Template: effects.TokenCard("2/2 white Cat Beast"), N: 1}.Apply(ctx)
			}),
		effects.ModeSteps("Put a +1/+1 counter on each creature you control. Those creatures gain vigilance until end of turn.", nil,
			// The vigilance waits for a prompt the counters ask (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.B11PutCounterOnEachCreatureYouControl(ctx.Game, ctx.Item)
			}),
			effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"vigilance"},
				Label:    "Felidar Retreat — vigilance",
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "16629f59-bae8-4c19-bf50-443eb0ed6856",
		Name:         "Felidar Retreat",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{landfall},
	})
}

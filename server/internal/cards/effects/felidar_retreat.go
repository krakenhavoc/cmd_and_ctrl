package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	landfall := Landfall("Felidar Retreat — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	landfall.Modes = ChooseOne(
		ModeDoing("Create a 2/2 white Cat Beast creature token.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Template: TokenCard("2/2 white Cat Beast"), N: 1}.Apply(ctx)
			}),
		ModeSteps("Put a +1/+1 counter on each creature you control. Those creatures gain vigilance until end of turn.", nil,
			// The vigilance waits for a prompt the counters ask (CR 608.2c, #4155).
			Instruction(func(ctx *Context) error {
				return b11PutCounterOnEachCreatureYouControl(ctx.Game, ctx.Item)
			}),
			GrantKeywordUntilEOT{
				Match:    And(Creature(), YouControl()),
				Keywords: []string{"vigilance"},
				Label:    "Felidar Retreat — vigilance",
			}),
	)
	Register(Spec{
		OracleID:     "16629f59-bae8-4c19-bf50-443eb0ed6856",
		Name:         "Felidar Retreat",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{landfall},
	})
}

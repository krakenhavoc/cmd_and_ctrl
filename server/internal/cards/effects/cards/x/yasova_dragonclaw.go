package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yasova Dragonclaw — {2}{G} Legendary Creature — Human Warrior 4/2 (Fate
// Reforged):
//
//	"Trample
//	 At the beginning of combat on your turn, you may pay {1}{U/R}{U/R}. If you
//	 do, gain control of target creature an opponent controls with power less
//	 than Yasova Dragonclaw's power until end of turn, untap that creature, and
//	 it gains haste until end of turn."
//
// Frenzied Goblin's shape: the target (an opponent's creature with power below
// Yasova's, compared with her power as she is now, or last known, CR 608.2h) is
// chosen as the trigger goes on the stack, before the payment is asked; the paid
// branch re-checks it as the ability resolves (CR 608.2b) and takes control
// until end of turn, untaps it and gives it haste, as Enthralling Victor does.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "75dc0995-f89e-4a2c-b30b-9b5a07c21183",
		Name:            "Yasova Dragonclaw",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtBeginningOfYourCombat("Yasova Dragonclaw — you may pay {1}{U/R}{U/R}", func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{U/R}{U/R}",
						Question: "Yasova Dragonclaw — pay {1}{U/R}{U/R} to gain control of the target creature until end of turn?",
						Then:     yasovaPaid,
					}.Apply(effects.NewContext(g, item))
				}),
				effects.RelativeToSource(effects.TargetCreature("target creature an opponent controls with power less than Yasova Dragonclaw's power", effects.OpponentControls()), effects.LesserPower())),
		},
	})
}

// yasovaPaid takes the chosen creature until end of turn, untaps it and gives
// it haste, if it is still a legal target.
var yasovaPaid = effects.PayUnlessBranches("pay-unless/yasova-dragonclaw-steal",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		id, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (effects.GainControl{Target: id, Duration: effects.DurationUntilEndOfTurn(ctx), Label: "Yasova Dragonclaw — gain control until end of turn"}).Apply(ctx); err != nil {
			return err
		}
		if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
		return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Yasova Dragonclaw — haste until end of turn"}.Apply(ctx)
	}, nil)

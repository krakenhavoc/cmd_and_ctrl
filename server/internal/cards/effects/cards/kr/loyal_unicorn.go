package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loyal Unicorn — Creature — Unicorn {3}{W}, 3/4:
//
//	"Vigilance
//	 Lieutenant — At the beginning of combat on your turn, if you
//	 control your commander, prevent all combat damage that would be
//	 dealt to creatures you control this turn. Other creatures you
//	 control gain vigilance until end of turn."
//
// Lieutenant is an ability word, so the condition is an intervening if
// (CR 603.4), Loyal Apprentice's: checked as the combat step begins and
// again as the trigger resolves (b18ControlsYourCommander, a commander
// you own and control).
//
// The shield is #2045's: creatures you control as the damage would be
// dealt (CR 611.2c), not you, combat damage only. The vigilance grant
// modifies characteristics, so it goes to the other creatures you
// control as the trigger resolves; one that enters later doesn't get it.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6cf8191c-acca-4145-86c2-53a138b3fe4a",
		Name:            "Loyal Unicorn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventStepBegan, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.StepBegan(game.StepBeginCombat, true)(ev, source, lki, g) &&
					effects.B18ControlsYourCommander(g, source.Controller)
			}, "Loyal Unicorn — prevent all combat damage to creatures you control; other creatures you control gain vigilance",
				loyalUnicornShield),
		},
	})
}

// loyalUnicornShield re-checks the lieutenant condition, then makes the
// shield and grants vigilance to the other creatures you control.
func loyalUnicornShield(g *game.Game, item *game.StackItem) error {
	if !effects.B18ControlsYourCommander(g, item.Controller) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	if err := (effects.PreventDamageFromSource{Protect: effects.ShieldCreaturesYouControl, CombatOnly: true}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{
		Match:    effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID)),
		Keywords: []string{"vigilance"},
		Label:    "Loyal Unicorn — other creatures you control gain vigilance",
	}.Apply(ctx)
}

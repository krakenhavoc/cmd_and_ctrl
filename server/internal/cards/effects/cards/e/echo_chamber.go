package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Echo Chamber — Artifact {4} (Tempest, S82 OLD b753, #5628):
//
//	"{4}, {T}: An opponent chooses target creature they control. Create a
//	 token that's a copy of that creature. That token gains haste until end
//	 of turn. Exile the token at the beginning of the next end step.
//	 Activate only as a sorcery."
//
// The target is an opponent's pick among the creatures they control
// (Preacher's TheyControl(OfAnOpponentsChoice)). The token is created under
// the Chamber's controller, as Rionya, Fire Dancer's copies are: haste rides
// the token's keywords, which is the same as "until end of turn" because the
// token is exiled at the next end step of the same turn (CR 603.7). The
// target is read again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75960649-93a4-4d56-b40f-f9fcbb4341d4",
		Name:         "Echo Chamber",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{4}, {T}: An opponent chooses target creature they control. Create a token that's a copy of that creature. That token gains haste until end of turn. Exile the token at the beginning of the next end step. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			SorcerySpeed: true,
			Targets: effects.TheyControl(effects.OfAnOpponentsChoice(
				effects.TargetCreature("target creature of an opponent's choice they control"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				target, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				tmpl, ok := effects.TokenCopyTemplate(g, target)
				if !ok {
					return nil
				}
				effects.TokenCopyGainsHaste(&tmpl)
				return g.CreateTokensKeyedForEffect(game.TokenCreation{
					Controller: ctx.Controller(),
					Source:     ctx.Source(),
					Groups:     []game.TokenGroup{{Template: effects.CopiedToken(tmpl), Count: 1}},
				}, ctx.ItemTail(echoChamberExileTheCopy))
			},
		}},
	})
}

// echoChamberExileTheCopy is "exile the token at the beginning of the next
// end step". An on-disk key.
var echoChamberExileTheCopy = effects.SocB8ExileTokensAtNextEndStep("echo-chamber/exile-the-copy",
	"Echo Chamber — exile the token")

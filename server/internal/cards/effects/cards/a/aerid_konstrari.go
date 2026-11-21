package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aerid Konstrari — Legendary Creature — Elder Sphinx {1}{R}{G}{G}, 5/4:
//
//	"Flying
//	 When Aerid Konstrari enters or dies, create a Heartwood token.
//	 (It's a red and green artifact with "{T}: Add {R} or {G}.")
//	 {6}: Create a Heartwood token. Then Aerid Konstrari gets +X/+0
//	 until end of turn, where X is the number of artifacts you
//	 control."
//
// X is counted after the token exists, so the new Heartwood is one of
// the artifacts, as the "Then" orders it. The pump is skipped if Aerid
// has left the battlefield by the time the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "172a8f58-c420-4497-b962-c5853f923667",
		Name:            "Aerid Konstrari",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrDies("Aerid Konstrari — create a Heartwood token",
				effects.Do(effects.CreateToken{Template: effects.HeartwoodToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}: Create a Heartwood token. Then Aerid Konstrari gets +X/+0 until end of turn, where X is the number of artifacts you control",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{6}"),
			// The boost waits for a prompt the token's creation asks (a CR 616 order; CR 608.2c, #4155)
			// and counts the artifacts when it runs.
			Effect: effects.Do(
				effects.CreateToken{Template: effects.HeartwoodToken(), N: 1},
				effects.Instruction(func(ctx *effects.Context) error {
					item := ctx.Item
					if !effects.B15OnBattlefield(ctx.Game, item.SourceCardID) {
						return nil
					}
					return effects.BoostUntilEOT{
						Target: item.SourceCardID,
						Power:  effects.RfCreatureAArtifactsYouControl(ctx.Game, item.Controller),
						Label:  "Aerid Konstrari — +X/+0",
					}.Apply(ctx)
				}),
			),
		}},
	})
}

package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iron Man, Modern Marvel — Legendary Artifact Creature — Human Hero
// {2}{U}{U}, 3/3 (Marvel Super Heroes Commander):
//
//	"Flying
//	 Other artifact creatures you control get +1/+1.
//	 Whenever Iron Man attacks, if you control another artifact creature,
//	 draw a card."
//
// Chief of the Foundry's lord (b27OtherArtifactCreaturesYouControl). The
// attack trigger is an intervening if (CR 603.4): checked as he attacks and
// again as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "507fc71d-0dfc-4bfe-bc71-de88162c39a1",
		Name:            "Iron Man, Modern Marvel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.B27OtherArtifactCreaturesYouControl, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control another artifact creature",
				Test: func(ctx *effects.Context) bool {
					for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Artifact(), effects.Creature(), effects.YouControl())) {
						if c.InstanceID != ctx.Source() {
							return true
						}
					}
					return false
				},
			}, effects.WheneverThisAttacks("Iron Man — draw a card", effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}

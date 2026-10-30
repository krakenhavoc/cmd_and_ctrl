package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "507fc71d-0dfc-4bfe-bc71-de88162c39a1",
		Name:            "Iron Man, Modern Marvel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			b16Anthem(b27OtherArtifactCreaturesYouControl, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control another artifact creature",
				Test: func(ctx *Context) bool {
					for _, c := range MatchingBattlefield(ctx, And(Artifact(), Creature(), YouControl())) {
						if c.InstanceID != ctx.Source() {
							return true
						}
					}
					return false
				},
			}, WheneverThisAttacks("Iron Man — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shimmer Dragon — Creature — Dragon {4}{U}{U}, 5/6 (Foundations
// Commander):
//
//	"Flying
//	 As long as you control four or more artifacts, this creature has
//	 hexproof.
//	 Tap two untapped artifacts you control: Draw a card."
//
// The hexproof is a layer 6 static on the Dragon alone, asked at every
// recompute (Wild Ogre Bupkis's artifact count; the Dragon is not an
// artifact itself). The draw is paid by tapping two untapped artifacts the
// activator controls (TapOthersCost, CR 118.3, 602.2b): not the {T} symbol,
// so a just-cast artifact may pay.
//
// No simplification.
func init() {
	const tapTwo = "two untapped artifacts you control"
	Register(Spec{
		OracleID:        "de1316f2-1cca-4ac4-aead-0fd8b4036d88",
		Name:            "Shimmer Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && stickerArtifactsControlledBy(g, source.Controller) >= 4
			}, "hexproof"),
		},
		Activated: []ActivatedAbility{{
			Label:   "Tap two untapped artifacts you control: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: TargetPermanent(tapTwo, Artifact()),
				Label:  tapTwo,
			}},
			Effect: b36DrawOne,
		}},
	})
}

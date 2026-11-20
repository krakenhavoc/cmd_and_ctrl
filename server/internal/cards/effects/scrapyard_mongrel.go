package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scrapyard Mongrel — Creature — Dog {3}{R}, 3/3 (Magic 2015, #5556):
//
//	"As long as you control an artifact, this creature gets +2/+0 and has
//	 trample."
//
// Gravblade Heavy's two conditional statics at +2/+0 and trample: a
// layer-7c pump and a layer-6 keyword on the Mongrel alone while its
// controller controls any artifact.
//
// No simplification.
func init() {
	whileArtifact := whileYouHave(cloudsculptControlsAnArtifact, selfCreature)
	trample := KeywordGrant(whileArtifact, "trample")
	trample.Label = "Has trample as long as you control an artifact."
	Register(Spec{
		OracleID:     "b6c8f2fe-b097-49af-ab59-16590dedaf32",
		Name:         "Scrapyard Mongrel",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Gets +2/+0 as long as you control an artifact.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: whileArtifact,
				Apply:     pumpBy(2, 0),
			},
			trample,
		},
	})
}

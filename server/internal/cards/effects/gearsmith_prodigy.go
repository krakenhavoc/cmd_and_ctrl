package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gearsmith Prodigy — Creature — Human Artificer {U}, 1/2
// (Core Set 2019, #5527):
//
//	"This creature gets +1/+0 as long as you control an artifact."
//
// Goblin Tomb Raider's +1/+0 while you control an artifact, without the
// haste.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "589d6b67-17ad-4ca7-9b0c-1b919bf03a27",
		Name:         "Gearsmith Prodigy",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Gets +1/+0 as long as you control an artifact.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: whileYouHave(cloudsculptControlsAnArtifact, selfCreature),
			Apply:     pumpBy(1, 0),
		}},
	})
}

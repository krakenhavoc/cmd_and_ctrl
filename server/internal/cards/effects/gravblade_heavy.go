package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gravblade Heavy — Creature — Human Soldier {3}{B}, 3/4 (Edge of
// Eternities, #4461):
//
//	"As long as you control an artifact, this creature gets +1/+0 and has
//	 deathtouch."
//
// Goblin Tomb Raider's shape with deathtouch in place of haste: one
// condition gating a layer 7c +1/+0 (CR 613.4c) and a layer 6 deathtouch
// grant (CR 613.1f) on the Heavy alone. "You control an artifact" is
// Cloudsculpt Technician's check, read after the layers, so a permanent
// made an artifact by another effect counts, and both end the moment the
// last artifact leaves.
//
// No simplification.
func init() {
	whileArtifact := whileYouHave(cloudsculptControlsAnArtifact, selfCreature)
	deathtouch := KeywordGrant(whileArtifact, "deathtouch")
	deathtouch.Label = "Has deathtouch as long as you control an artifact."
	Register(Spec{
		OracleID:     "67821712-e58f-4fee-81af-021059ac72ea",
		Name:         "Gravblade Heavy",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Gets +1/+0 as long as you control an artifact.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: whileArtifact,
				Apply:     pumpBy(1, 0),
			},
			deathtouch,
		},
	})
}

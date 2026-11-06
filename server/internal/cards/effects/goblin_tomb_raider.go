package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Tomb Raider — Creature — Goblin Pirate {R}, 1/2 (The Lost
// Caverns of Ixalan, #4470):
//
//	"As long as you control an artifact, this creature gets +1/+0 and has
//	 haste."
//
// One condition gating two statics on the Raider alone: a layer 7c +1/+0
// (CR 613.4c) and a layer 6 haste grant (CR 613.1f). "You control an
// artifact" is Cloudsculpt Technician's check, read after the layers, so a
// permanent made an artifact by another effect counts, and both end the
// moment the last artifact leaves. Haste (CR 702.10a) lets it attack the
// turn it arrives if an artifact is there by the declare attackers step.
//
// No simplification.
func init() {
	whileArtifact := whileYouHave(cloudsculptControlsAnArtifact, selfCreature)
	haste := KeywordGrant(whileArtifact, "haste")
	haste.Label = "Has haste as long as you control an artifact."
	Register(Spec{
		OracleID:     "37ce95bb-8ff5-47b9-85e4-21be7a794940",
		Name:         "Goblin Tomb Raider",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Gets +1/+0 as long as you control an artifact.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: whileArtifact,
				Apply:     pumpBy(1, 0),
			},
			haste,
		},
	})
}

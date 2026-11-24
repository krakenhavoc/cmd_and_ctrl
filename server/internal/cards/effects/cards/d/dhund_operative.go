package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dhund Operative — Creature — Human Rogue {1}{B}, 2/2 (Commander Legends):
//
//	"As long as you control an artifact, this creature gets +1/+0 and has
//	 deathtouch."
//
// Goblin Tomb Raider's pair of statics over one condition, with deathtouch in
// place of haste: a layer 7c +1/+0 (CR 613.4c) and a layer 6 keyword grant
// (CR 613.1f), both read after the layers so a permanent made an artifact by
// another effect counts, and both end the moment the last artifact leaves.
//
// No simplification.
func init() {
	whileArtifact := effects.WhileYouHave(effects.CloudsculptControlsAnArtifact, effects.SelfCreature)
	deathtouch := effects.KeywordGrant(whileArtifact, "deathtouch")
	deathtouch.Label = "Has deathtouch as long as you control an artifact."
	effects.Register(effects.Spec{
		OracleID:     "e025c802-9862-4dee-8c2a-b875eb001714",
		Name:         "Dhund Operative",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Gets +1/+0 as long as you control an artifact.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: whileArtifact,
				Apply:     effects.PumpBy(1, 0),
			},
			deathtouch,
		},
	})
}

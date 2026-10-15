package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Girder Goons — Creature — Ogre Warrior {4}{B}, 4/4:
//
//	"When this creature dies, create a tapped 2/2 black Rogue creature
//	 token.
//	 Blitz {3}{B}"
//
// Blitz (CR 702.152, #3772). The Rogue enters tapped: the template
// carries Tapped, so it never enters untapped and taps a beat later.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "655961a3-94b5-4600-b9f1-f93163d03a49",
		Name:             "Girder Goons",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Blitz("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Girder Goons — create a tapped 2/2 black Rogue",
				Do(CreateToken{Template: girderGoonsRogue(), N: 1})),
		},
	})
}

// girderGoonsRogue is the 2/2 black Rogue, tapped.
func girderGoonsRogue() game.Card {
	tmpl := TokenCard("2/2 black Rogue")
	tmpl.Tapped = true
	return tmpl
}

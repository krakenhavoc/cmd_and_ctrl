package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jace's Mindseeker — Creature — Fish Illusion {4}{U}{U}, 4/4 (Magic 2014, S82
// old sets):
//
//	"Flying
//	 When this creature enters, target opponent mills five cards. You may cast
//	 an instant or sorcery spell from among them without paying its mana cost."
//
// Sorcerous Squall's mill-then-cast, narrowed to the cards this mill put into the
// graveyard ("from among them", not the rest of the graveyard) and without the
// exile-instead rider the Squall prints. The mill finishes first (CR 608.2c),
// the cast is made during the resolution (CR 608.2g) through the resolution-cast
// prompt under one free, cast-only permission, so the first cast spends it: one
// spell. A pass closes the window and leaves the card where it is. A target that
// is gone as the ability resolves does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ebf337d-eac0-40de-978c-3a8fc3bbe03c",
		Name:            "Jace's Mindseeker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Jace's Mindseeker — target opponent mills five cards; you may cast an instant or sorcery spell from among them without paying its mana cost",
				effects.S82OldB400MindseekerMill),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

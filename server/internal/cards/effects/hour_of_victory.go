package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hour of Victory — Enchantment {2}{B}:
//
//	"Start your engines!
//	 When this enchantment enters, create a 2/2 black Zombie creature
//	 token.
//	 Max speed — {1}{B}, Sacrifice this enchantment: Search your library
//	 for a card, put it into your hand, then shuffle. Activate only as a
//	 sorcery."
//
// ADR 0138 (#2122, #2711). The tutor is a max-speed activated ability
// (greyed below max speed) that is also sorcery speed, so both gates hold.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eb100b3b-bc3f-48f3-8f28-bb2d80225ea9",
		Name:            "Hour of Victory",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hour of Victory — create a 2/2 black Zombie creature token",
				Do(CreateToken{Template: BlackZombieToken(), N: 1})),
		},
		Activated: []ActivatedAbility{
			MaxSpeedActivated(ActivatedAbility{
				Label:        "Max speed — {1}{B}, Sacrifice this enchantment: Search your library for a card, put it into your hand, then shuffle. Activate only as a sorcery.",
				Purpose:      game.Purpose{Tutors: 1},
				Cost:         Plus(ManaCost("{1}{B}"), SacrificeThis()),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Source:    item.SourceCardID,
						Predicate: func(game.Card) bool { return true },
						Dest:      game.ZoneHand,
						Limit:     1,
						Shuffle:   true,
						Reason:    "Hour of Victory — search for a card",
					}.Apply(NewContext(g, item))
				},
			}),
		},
	})
}

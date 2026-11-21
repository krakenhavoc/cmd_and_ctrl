package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "eb100b3b-bc3f-48f3-8f28-bb2d80225ea9",
		Name:            "Hour of Victory",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hour of Victory — create a 2/2 black Zombie creature token",
				effects.Do(effects.CreateToken{Template: effects.BlackZombieToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			effects.MaxSpeedActivated(effects.ActivatedAbility{
				Label:        "Max speed — {1}{B}, Sacrifice this enchantment: Search your library for a card, put it into your hand, then shuffle. Activate only as a sorcery.",
				Purpose:      game.Purpose{Tutors: 1},
				Cost:         effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeThis()),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:  item.Controller,
						Source:  item.SourceCardID,
						Dest:    game.ZoneHand,
						Limit:   1,
						Shuffle: true,
						Reason:  "Hour of Victory — search for a card",
					}.Apply(effects.NewContext(g, item))
				},
			}),
		},
	})
}

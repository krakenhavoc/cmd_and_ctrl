package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sultai Monument — Artifact {2} (Tarkir: Dragonstorm, #4463):
//
//	"When this artifact enters, search your library for a basic Swamp,
//	 Forest, or Island card, reveal it, put it into your hand, then
//	 shuffle.
//	 {2}{B}{G}{U}, {T}, Sacrifice this artifact: Create two 2/2 black
//	 Zombie Druid creature tokens. Activate only as a sorcery."
//
// Abzan Monument's shape: the entry search is the basic-of-three
// predicate (IsBasicLandOfAnySubtype) sent to the hand and revealed; the
// activated ability is sorcery speed, and the Monument is sacrificed as
// part of its cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9db71b76-eda7-49dd-af5a-3de18acb8ccb",
		Name:         "Sultai Monument",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sultai Monument — search for a basic Swamp, Forest, or Island card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLandOfAnySubtype("Swamp", "Forest", "Island"),
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Sultai Monument — a basic Swamp, Forest, or Island card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{B}{G}{U}, {T}, Sacrifice this artifact: Create two 2/2 black Zombie Druid creature tokens. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}{B}{G}{U}"), effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie Druid"), N: 2}),
		}},
	})
}

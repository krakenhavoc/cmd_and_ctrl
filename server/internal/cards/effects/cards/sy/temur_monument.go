package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Temur Monument — Artifact {2} (Tarkir: Dragonstorm, #4463):
//
//	"When this artifact enters, search your library for a basic Forest,
//	 Island, or Mountain card, reveal it, put it into your hand, then
//	 shuffle.
//	 {3}{G}{U}{R}, {T}, Sacrifice this artifact: Create a 5/5 green
//	 Elephant creature token. Activate only as a sorcery."
//
// Abzan Monument's shape: the entry search is the basic-of-three
// predicate (IsBasicLandOfAnySubtype) sent to the hand and revealed; the
// activated ability is sorcery speed, and the Monument is sacrificed as
// part of its cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5464d5ae-5be5-4e52-b9d0-ddce48fd93b3",
		Name:         "Temur Monument",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Temur Monument — search for a basic Forest, Island, or Mountain card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLandOfAnySubtype("Forest", "Island", "Mountain"),
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Temur Monument — a basic Forest, Island, or Mountain card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{G}{U}{R}, {T}, Sacrifice this artifact: Create a 5/5 green Elephant creature token. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{3}{G}{U}{R}"), effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("5/5 green Elephant"), N: 1}),
		}},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5464d5ae-5be5-4e52-b9d0-ddce48fd93b3",
		Name:         "Temur Monument",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Temur Monument — search for a basic Forest, Island, or Mountain card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLandOfAnySubtype("Forest", "Island", "Mountain"),
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Temur Monument — a basic Forest, Island, or Mountain card",
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label:        "{3}{G}{U}{R}, {T}, Sacrifice this artifact: Create a 5/5 green Elephant creature token. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}{G}{U}{R}"), TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			Effect:       Do(CreateToken{Template: TokenCard("5/5 green Elephant"), N: 1}),
		}},
	})
}

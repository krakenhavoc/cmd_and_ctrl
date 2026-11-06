package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9db71b76-eda7-49dd-af5a-3de18acb8ccb",
		Name:         "Sultai Monument",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sultai Monument — search for a basic Swamp, Forest, or Island card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLandOfAnySubtype("Swamp", "Forest", "Island"),
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Sultai Monument — a basic Swamp, Forest, or Island card",
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label:        "{2}{B}{G}{U}, {T}, Sacrifice this artifact: Create two 2/2 black Zombie Druid creature tokens. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}{B}{G}{U}"), TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			Effect:       Do(CreateToken{Template: TokenCard("2/2 black Zombie Druid"), N: 2}),
		}},
	})
}

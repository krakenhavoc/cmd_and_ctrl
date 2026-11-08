package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jeskai Monument — Artifact {2} (Tarkir: Dragonstorm, #4463):
//
//	"When this artifact enters, search your library for a basic Island,
//	 Mountain, or Plains card, reveal it, put it into your hand, then
//	 shuffle.
//	 {1}{U}{R}{W}, {T}, Sacrifice this artifact: Create two 1/1 white
//	 Bird creature tokens with flying. Activate only as a sorcery."
//
// Abzan Monument's shape. The entry search is the basic-of-three
// predicate (IsBasicLandOfAnySubtype) sent to the hand and revealed. The
// Birds' flying is part of the token's definition, so it lasts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0b2b7d04-73f3-43f6-86e0-fa00cbbc6469",
		Name:         "Jeskai Monument",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Jeskai Monument — search for a basic Island, Mountain, or Plains card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLandOfAnySubtype("Island", "Mountain", "Plains"),
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Jeskai Monument — a basic Island, Mountain, or Plains card",
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label: "{1}{U}{R}{W}, {T}, Sacrifice this artifact: Create two 1/1 white Bird creature tokens with flying. " +
				"Activate only as a sorcery.",
			Purpose:      game.Purpose{Tokens: 2},
			Cost:         Plus(ManaCost("{1}{U}{R}{W}"), TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			Effect:       Do(CreateToken{Template: TokenCard("1/1 white Bird with flying"), N: 2}),
		}},
	})
}

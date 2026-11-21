package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0b2b7d04-73f3-43f6-86e0-fa00cbbc6469",
		Name:         "Jeskai Monument",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Jeskai Monument — search for a basic Island, Mountain, or Plains card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLandOfAnySubtype("Island", "Mountain", "Plains"),
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Jeskai Monument — a basic Island, Mountain, or Plains card",
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{U}{R}{W}, {T}, Sacrifice this artifact: Create two 1/1 white Bird creature tokens with flying. " +
				"Activate only as a sorcery.",
			Purpose:      game.Purpose{Tokens: 2},
			Cost:         effects.Plus(effects.ManaCost("{1}{U}{R}{W}"), effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Bird with flying"), N: 2}),
		}},
	})
}

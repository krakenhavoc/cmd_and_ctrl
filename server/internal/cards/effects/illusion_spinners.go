package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Illusion Spinners — Creature — Faerie Wizard {4}{U}, 4/3 (Lorwyn
// Eclipsed, #4458):
//
//	"You may cast this spell as though it had flash if you control a
//	 Faerie.
//	 Flying
//	 This creature has hexproof as long as it's untapped."
//
// The flash clause is the card's own timing permission (Bard's Company's
// FlashIfYou), its condition read as the cast begins (CR 601.3): a Faerie
// you control then, a changeling counting. Hexproof is a layer 6 grant to
// the creature itself that applies only while it is untapped, re-read on
// every layer pass, so tapping it (attacking without vigilance) removes
// hexproof at once and untapping restores it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b608288e-1755-4738-b828-9df6998bae8d",
		Name:            "Illusion Spinners",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		FlashIf:         FlashIfYou("if you control a Faerie", illusionSpinnersYouControlAFaerie),
		Static: []game.StaticAbility{
			func() game.StaticAbility {
				ab := KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID && !target.Tapped
				}, "hexproof")
				ab.Label = "This creature has hexproof as long as it's untapped."
				return ab
			}(),
		},
	})
}

func illusionSpinnersYouControlAFaerie(g *game.Game, you uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.HasSubtype("Faerie") {
			return true
		}
	}
	return false
}

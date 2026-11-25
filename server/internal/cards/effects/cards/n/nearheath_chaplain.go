package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nearheath Chaplain — Creature — Human Cleric {3}{W}, 3/1 (Shadows over
// Innistrad, #5543):
//
//	"Lifelink
//	 {2}{W}, Exile this card from your graveyard: Create two 1/1 white
//	 Spirit creature tokens with flying. Activate only as a sorcery."
//
// A graveyard activation (ActivatedAbilityShape.Zones) whose cost exiles the
// card itself (ExileThis), at sorcery speed (CR 602.5d). The tokens are made
// even though the card is already gone, since the exile is a cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "16c587db-d07b-47ab-8b03-c3b2cc7144f5",
		Name:            "Nearheath Chaplain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{W}, Exile this card from your graveyard: Create two 1/1 white Spirit creature tokens with flying. Activate only as a sorcery.",
			Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
			Cost:         effects.Plus(effects.ManaCost("{2}{W}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 2}),
		}},
	})
}
